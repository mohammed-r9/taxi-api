package network

func (api *Api) Init() {
	api.Engine.HTTPErrorHandler = ErrorHandler(api.Logger)
	api.OpenAPI = newOpenAPIDocument()
	api.mountDocs()
}
