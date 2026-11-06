package network

import (
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/labstack/echo/v5"
)

const (
	apiTitle            = "Taxi API"
	apiVersion          = "0.1.0"
	docsPath            = "/api/docs"
	openAPIDocumentPath = "/api/docs/openapi.json"
	openAPIAliasPath    = "/api/docs/openapi"
)

func newOpenAPIDocument() *openapi3.T {
	return &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   apiTitle,
			Version: apiVersion,
		},
		Paths: openapi3.NewPaths(),
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{},
			SecuritySchemes: openapi3.SecuritySchemes{
				"BearerAuth": {
					Value: openapi3.NewSecurityScheme().
						WithType("http").
						WithScheme("bearer").
						WithBearerFormat("JWT").
						WithDescription("JWT access token. The Bearer prefix is added automatically."),
				},
			},
		},
	}
}

func (api *Api) mountDocs() {
	api.Engine.GET(docsPath, api.serveDocs)
	api.Engine.GET(openAPIDocumentPath, api.serveOpenAPI)
	api.Engine.GET(openAPIAliasPath, api.serveOpenAPI)
}

func (api *Api) serveDocs(c *echo.Context) error {
	return c.HTML(http.StatusOK, scalarDocsHTML)
}

func (api *Api) serveOpenAPI(c *echo.Context) error {
	document, err := api.OpenAPI.MarshalJSON()
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, "application/openapi+json", document)
}

const scalarDocsHTML = `<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Taxi API</title>
</head>
<body>
	<script id="api-reference" data-url="/api/docs/openapi.json"></script>
	<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`

type openapiOperation struct {
	ID          string
	Summary     string
	Description string
	Tag         string
	Path        string
	ContentType string
	Errors      []int
	Auth        bool
	Permissions []string
	Status      int
	Input       reflect.Type
	Output      reflect.Type
	Schemas     openapi3.Schemas
}

func (o *openapiOperation) build() *openapi3.Operation {
	operation := openapi3.NewOperation()
	operation.OperationID = o.ID
	operation.Summary = o.Summary
	operation.Description = o.Description
	operation.Responses = openapi3.NewResponses()

	if o.Tag != "" {
		operation.Tags = []string{o.Tag}
	}
	if o.Auth {
		operation.Security = &openapi3.SecurityRequirements{{"BearerAuth": {}}}
	}
	if len(o.Permissions) > 0 {
		required := "Required permissions: `" + strings.Join(o.Permissions, "`, `") + "`"
		if operation.Description == "" {
			operation.Description = required
		} else {
			operation.Description += "\n\n" + required
		}
	}

	if bodyType, ok := requestBodyType(o.Input); ok {
		operation.RequestBody = &openapi3.RequestBodyRef{
			Value: openapi3.NewRequestBody().
				WithRequired(true).
				WithJSONSchemaRef(o.bodySchemaRef(bodyType)),
		}
	}

	o.addParameters(operation, "path", o.Input)
	o.addPathPlaceholders(operation)
	o.addParameters(operation, "query", o.Input)
	o.addParameters(operation, "header", o.Input)

	status := o.Status
	if status == 0 {
		status = http.StatusOK
	}
	response := openapi3.NewResponse().WithDescription(http.StatusText(status))
	if bodyType, ok := responseBodyType(o.Output); ok {
		contentType := o.ContentType
		if contentType == "" {
			contentType = "application/json"
		}
		response.WithContent(openapi3.NewContentWithSchemaRef(o.schemaRef(bodyType), []string{contentType}))
	}
	operation.Responses.Set(strconv.Itoa(status), &openapi3.ResponseRef{Value: response})

	errorSchema := o.errorSchemaRef()
	operation.Responses.Set("default", newErrorResponse(errorSchema, "Error response"))
	for _, code := range o.Errors {
		operation.Responses.Set(strconv.Itoa(code), newErrorResponse(errorSchema, http.StatusText(code)))
	}

	return operation
}

func (o *openapiOperation) errorSchemaRef() *openapi3.SchemaRef {
	const name = "ErrorResponse"

	ref := "#/components/schemas/" + name
	if _, ok := o.Schemas[name]; !ok {
		o.Schemas[name] = o.schemaRef(reflect.TypeFor[ErrorResponse]())
	}

	return &openapi3.SchemaRef{Ref: ref}
}

func newErrorResponse(schemaRef *openapi3.SchemaRef, description string) *openapi3.ResponseRef {
	content := openapi3.NewContentWithJSONSchemaRef(schemaRef)
	content["application/json"].Example = ErrorResponse{
		Success: false,
		Status:  http.StatusBadRequest,
		Code:    "VALIDATION_ERROR",
		Message: "request validation failed",
		Detail:  []string{"field is required"},
	}

	return &openapi3.ResponseRef{
		Value: openapi3.NewResponse().
			WithDescription(description).
			WithContent(content),
	}
}

func (o *openapiOperation) addParameters(operation *openapi3.Operation, in string, input reflect.Type) {
	input = indirectType(input)
	if input.Kind() != reflect.Struct {
		return
	}

	tag := parameterTag(in)
	for i := range input.NumField() {
		param := input.Field(i)
		if param.PkgPath != "" {
			continue
		}

		value, ok := param.Tag.Lookup(tag)
		if !ok {
			continue
		}
		name, _, _ := strings.Cut(value, ",")
		if name == "" || name == "-" {
			continue
		}

		operation.AddParameter(&openapi3.Parameter{
			Name:        name,
			In:          in,
			Description: param.Tag.Get("description"),
			Required:    in == "path" || isRequired(param),
			Schema:      o.schemaRef(indirectType(param.Type)),
		})
	}
}

func (o *openapiOperation) addPathPlaceholders(operation *openapi3.Operation) {
	existing := make(map[string]bool)
	for _, parameter := range operation.Parameters {
		if parameter.Value != nil && parameter.Value.In == "path" {
			existing[parameter.Value.Name] = true
		}
	}

	for _, part := range strings.Split(toOpenAPIPath(o.Path), "/") {
		name, ok := strings.CutPrefix(part, "{")
		if !ok {
			continue
		}
		name, ok = strings.CutSuffix(name, "}")
		if !ok || name == "" || existing[name] {
			continue
		}

		operation.AddParameter(&openapi3.Parameter{
			Name:     name,
			In:       "path",
			Required: true,
			Schema:   o.schemaRef(reflect.TypeFor[string]()),
		})
		existing[name] = true
	}
}

func (o *openapiOperation) schemaRef(t reflect.Type) *openapi3.SchemaRef {
	if o.Schemas == nil {
		o.Schemas = openapi3.Schemas{}
	}

	ref, err := openapi3gen.NewSchemaRefForValue(
		reflect.Zero(t).Interface(),
		o.Schemas,
		openapi3gen.SchemaCustomizer(applySchemaTags),
	)
	if err != nil {
		panic(fmt.Sprintf("openapi: generate schema for %s: %v", t, err))
	}

	markRequired(t, ref, o.Schemas, map[reflect.Type]bool{})
	return ref
}

func (o *openapiOperation) bodySchemaRef(t reflect.Type) *openapi3.SchemaRef {
	ref := o.schemaRef(t)
	excluded := excludedBodyFields(t)
	if len(excluded) == 0 {
		return ref
	}

	schema := resolveSchema(ref, o.Schemas)
	if schema == nil {
		return ref
	}

	for name := range excluded {
		delete(schema.Properties, name)
	}
	schema.Required = slices.DeleteFunc(schema.Required, func(name string) bool {
		return excluded[name]
	})
	return ref
}

func requestBodyType(t reflect.Type) (reflect.Type, bool) {
	t = indirectType(t)
	if t.Kind() != reflect.Struct {
		return t, true
	}
	if t.NumField() == 0 {
		return nil, false
	}

	excluded := excludedBodyFields(t)
	for i := range t.NumField() {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := jsonName(field)
		if name == "" || excluded[name] {
			continue
		}
		return t, true
	}
	return nil, false
}

func responseBodyType(t reflect.Type) (reflect.Type, bool) {
	t = indirectType(t)
	if t.Kind() == reflect.Struct && t.NumField() == 0 {
		return nil, false
	}
	return t, true
}

func excludedBodyFields(t reflect.Type) map[string]bool {
	t = indirectType(t)
	excluded := make(map[string]bool)
	if t.Kind() != reflect.Struct {
		return excluded
	}

	for i := range t.NumField() {
		field := t.Field(i)
		if field.PkgPath != "" || !hasBindingTag(field) {
			continue
		}
		if name := jsonName(field); name != "" {
			excluded[name] = true
		}
	}
	return excluded
}

func hasBindingTag(field reflect.StructField) bool {
	for _, tag := range []string{"param", "query", "header", "form"} {
		if _, ok := field.Tag.Lookup(tag); ok {
			return true
		}
	}
	return false
}

func parameterTag(in string) string {
	if in == "path" {
		return "param"
	}
	return in
}

func indirectType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func toOpenAPIPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + strings.TrimPrefix(part, ":") + "}"
		}
	}
	return strings.Join(parts, "/")
}

func applySchemaTags(_ string, _ reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	if value := tag.Get("description"); value != "" {
		schema.Description = value
	}
	if value := tag.Get("format"); value != "" {
		schema.Format = value
	}
	if value := tag.Get("pattern"); value != "" {
		schema.Pattern = value
	}
	if value := tag.Get("minLength"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		schema.MinLength = parsed
	}
	if value := tag.Get("maxLength"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		schema.MaxLength = &parsed
	}
	if value := tag.Get("minimum"); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		schema.Min = &parsed
	}
	if value := tag.Get("maximum"); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		schema.Max = &parsed
	}
	if value := tag.Get("minItems"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		schema.MinItems = parsed
	}
	if value := tag.Get("maxItems"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		schema.MaxItems = &parsed
	}
	if value := tag.Get("example"); value != "" {
		schema.Example = tagValue(value, schema)
	}
	if value := tag.Get("enum"); value != "" {
		for _, item := range strings.Split(value, ",") {
			schema.Enum = append(schema.Enum, tagValue(item, schema))
		}
	}
	if tag.Get("deprecated") == "true" {
		schema.Deprecated = true
	}
	if tag.Get("readOnly") == "true" {
		schema.ReadOnly = true
	}
	if tag.Get("writeOnly") == "true" {
		schema.WriteOnly = true
	}
	return nil
}

func tagValue(value string, schema *openapi3.Schema) any {
	if schema.Type != nil {
		switch {
		case schema.Type.Is("integer"):
			if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
				return parsed
			}
		case schema.Type.Is("number"):
			if parsed, err := strconv.ParseFloat(value, 64); err == nil {
				return parsed
			}
		case schema.Type.Is("boolean"):
			if parsed, err := strconv.ParseBool(value); err == nil {
				return parsed
			}
		}
	}
	return value
}

func markRequired(t reflect.Type, ref *openapi3.SchemaRef, schemas openapi3.Schemas, seen map[reflect.Type]bool) {
	t = indirectType(t)
	if seen[t] {
		return
	}
	seen[t] = true

	schema := resolveSchema(ref, schemas)
	if schema == nil {
		return
	}

	switch t.Kind() {
	case reflect.Struct:
		for i := range t.NumField() {
			field := t.Field(i)
			if field.PkgPath != "" {
				continue
			}

			name := jsonName(field)
			if name == "" {
				continue
			}

			if isRequired(field) && !slices.Contains(schema.Required, name) {
				schema.Required = append(schema.Required, name)
			}
			if property, ok := schema.Properties[name]; ok {
				markRequired(field.Type, property, schemas, seen)
			}
		}
	case reflect.Slice, reflect.Array:
		markRequired(t.Elem(), schema.Items, schemas, seen)
	case reflect.Map:
		markRequired(t.Elem(), schema.AdditionalProperties.Schema, schemas, seen)
	}
}

func resolveSchema(ref *openapi3.SchemaRef, schemas openapi3.Schemas) *openapi3.Schema {
	if ref == nil {
		return nil
	}
	if ref.Value != nil {
		return ref.Value
	}
	if !strings.HasPrefix(ref.Ref, "#/components/schemas/") {
		return nil
	}
	name := strings.TrimPrefix(ref.Ref, "#/components/schemas/")
	if schema, ok := schemas[name]; ok {
		return resolveSchema(schema, schemas)
	}
	return nil
}

func jsonName(field reflect.StructField) string {
	value, ok := field.Tag.Lookup("json")
	if !ok {
		return ""
	}

	name, _, _ := strings.Cut(value, ",")
	switch name {
	case "-":
		return ""
	case "":
		return field.Name
	default:
		return name
	}
}

func isRequired(field reflect.StructField) bool {
	if field.Tag.Get("required") == "true" {
		return true
	}
	for part := range strings.SplitSeq(field.Tag.Get("validate"), ",") {
		if strings.TrimSpace(part) == "required" {
			return true
		}
	}
	return false
}
