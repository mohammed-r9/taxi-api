package network

func (api *Api) Init() {
	api.Engine.HTTPErrorHandler = ErrorHandler(api.Engine.Logger)
	api.OpenAPI = newOpenAPIDocument()
	api.mountDocs()
}
