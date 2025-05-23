package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/JackBekket/reflexia/internal/api"
	"github.com/JackBekket/reflexia/pkg/config"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/swaggest/openapi-go/openapi31"
	"github.com/swaggest/rest/nethttp"
	"github.com/swaggest/rest/response/gzip"
	"github.com/swaggest/rest/web"
	swgui "github.com/swaggest/swgui/v5emb"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
)

func Run(cfg config.Config) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	workdir, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("get workdir")
	}

	listenAddr := getEnv("LISTEN_ADDR")
	allowedOriginsStr := getEnv("CORS_ALLOW_ORIGINS")
	allowedOrigins := strings.Split(
		strings.TrimSpace(
			allowedOriginsStr,
		), ",")

	webService := web.NewService(openapi31.NewReflector())

	webService.OpenAPISchema().SetTitle("Reflexia API")
	webService.OpenAPISchema().SetVersion("v0.0.1")

	webService.Use(
		cors.Handler(
			cors.Options{
				AllowedOrigins: allowedOrigins,
				AllowedHeaders: []string{
					"content-type",
				},
			},
		),
	)

	webService.Wrap(
		gzip.Middleware,
		middleware.Logger,
		middleware.Recoverer,
	)

	apiService := api.APIService{
		Workdir: workdir,
		Config:  cfg,
	}

	projectConfigsInteractor := usecase.NewInteractor(apiService.ProjectConfigsGet)
	projectConfigsInteractor.SetTitle("Project configs")
	projectConfigsInteractor.SetDescription(
		"Can be used to retrieve available .toml project config prompt templates " +
			"with project guessing info (project root directory defining files, filetypes)",
	)
	projectConfigsInteractor.SetExpectedErrors(
		status.Internal,
	)
	webService.Method(http.MethodGet, "/project_configs", nethttp.NewHandler(projectConfigsInteractor))

	reflectInteractor := usecase.NewInteractor(apiService.ReflectPost)
	reflectInteractor.SetTitle("Reflect")
	reflectInteractor.SetDescription(
		"The main function of that service. Given github repository URL clones it, " +
			"then summarizing it per file and per package, producing README.md in each package directory.\n" +
			"If PR creating set - commits the result to [branch_name]_autodoc branch, force pushes it, " +
			"and returning github PR URL.",
	)
	reflectInteractor.SetExpectedErrors(
		status.Internal,
		status.InvalidArgument,
	)

	webService.Method(http.MethodPost, "/reflect", nethttp.NewHandler(reflectInteractor))

	webService.Docs("/docs", swgui.New)

	if err := http.ListenAndServe(listenAddr, webService); err != nil {
		log.Fatal().Err(err).Msgf("serving at: %s", listenAddr)
	}
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatal().Msgf("%s can't be empty", key)
	}
	return value
}
