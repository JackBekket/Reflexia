# reflexia/api

The `reflexia/api` package provides an HTTP server for handling API endpoints, including project configuration management and GitHub repository operations. It integrates with OpenAPI for documentation and supports configuration via environment variables and a config struct.

**File Structure**  
- `api.go`: Contains the main API server implementation.

**External Configuration**  
- **Environment Variables**:  
  - `LISTEN_ADDR` (string): Server listen address (required)  
  - `CORS_ALLOW_ORIGINS` (string): Comma-separated list of allowed CORS origins (required)  
- **Config Struct**: `pkg/config.Config` (application-wide configuration)

**Behavior**  
- Initializes an HTTP server with logging, middleware, and API handlers.  
- Registers endpoints:  
  - `/project_configs`: Project configuration management.  
  - `/reflect`: GitHub repository operations (error codes: `Internal`, `InvalidArgument`).  
  - `/docs`: OpenAPI documentation via `swgui` (no custom documentation provided).  

**Limitations/Notes**  
- API endpoints `/project_configs` and `/reflect` are not fully documented in OpenAPI.  
- No error handling for invalid OpenAPI routes.  
- Workdir is determined at startup and used by the APIService.  
- Environment variables `LISTEN_ADDR` and `CORS_ALLOW_ORIGINS` are critical; missing values cause panic.  

**Dependencies**  
Uses `go-chi/chi/v5`, `swaggest/openapi-go`, and `rs/zerolog` for routing, OpenAPI integration, and logging.