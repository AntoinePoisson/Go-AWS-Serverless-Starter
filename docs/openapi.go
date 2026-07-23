// Package docs carries the OpenAPI general information and nothing else: it
// declares no Go symbol and no package imports it. Run `make docs` after
// touching an annotation.
//
// It sits outside both functions because `swag init -g` accepts exactly one
// general-information file, and neither binary owns the title of the service.
package docs

//	@title			bootstrap-go-aws API
//	@version		1.0
//	@description	Item resource served by two Lambda functions behind a single HTTP API.
//	@description
//	@description	The /items routes are authenticated with the X-Api-Key header and are
//	@description	served by the `api` function. /health and /public/... are open and are
//	@description	served by the `public` function. Both sit behind the same API Gateway,
//	@description	which is why the public read is namespaced under /public.
//	@description
//	@description				Errors carry a JSON body of the shape {"code": "...", "message": "..."},
//	@description				except for the 404 and 405 produced by the router itself.
//
//	@externalDocs.url			https://github.com/antoinepoisson/bootstrap-go-aws
//	@externalDocs.description	Source repository and README
//
//	@servers.url				http://localhost:8080
//	@servers.description		Local api function
//
//	@servers.url				http://localhost:8081
//	@servers.description		Local public function
//
//	@servers.url				https://{apiId}.execute-api.{region}.amazonaws.com
//	@servers.description		Deployed stage, both functions behind one API Gateway
//	@servers.variables.default	apiId	xxxxxxxxxx
//	@servers.variables.default	region	eu-west-1
//
// The tags must come BEFORE @securityDefinitions: its description swallows
// every annotation that follows, and anything declared after it vanishes from
// the generated document without a warning.
//
//	@tag.name					items
//	@tag.description			Authenticated CRUD over the item resource.
//	@tag.name					public
//	@tag.description			Unauthenticated reads, namespaced under /public.
//	@tag.name					health
//	@tag.description			Liveness, stage and version.
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						X-Api-Key
//	@description				Shared secret. Deployed stages read it from SSM Parameter Store.
//
// The version above is the API contract, not a release: Release Please tags the
// repository and never touches this file. Bump it when a change breaks a
// consumer.
//
// Never open a prose line with an at-sign - the swaggo formatter reads it as an
// annotation and folds it into the block above.
