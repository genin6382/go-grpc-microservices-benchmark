package gateway 

import (

)

// Centralised router for the gateway. This will be used to route requests to the appropriate services based on the request path and method. It will also handle authentication and authorization for the requests.
type Router struct {
	// Add fields for the router here, such as a map of routes to handlers, a logger, etc.
}

// NewRouter creates a new instance of the Router.
func NewRouter() *Router {
	return &Router{
		// Initialize fields here, such as the routes map, logger, etc.
	}
}

// RouteRequest routes the incoming request to the appropriate handler based on the request path and method. It also handles authentication and authorization for the request.
