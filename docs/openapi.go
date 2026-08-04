// Package docs carries the OpenAPI general information and nothing else. It
// declares no Go symbol and no package imports it.
//
// swag reads the annotations below, plus the ones on each handler method, and
// writes docs/openapi.yaml — the file Redocly renders and the one reviewers
// read. Regenerate with `make docs` after touching any annotation.
//
// It lives outside lambda/api and lambda/public on purpose. `swag init -g`
// accepts exactly one general-information file, and putting the title of the
// whole service inside one of the two binaries would make that choice arbitrary
// and the other function a second-class citizen of its own documentation.
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
// The tags come BEFORE @securityDefinitions, and the order is load-bearing:
// the @description that belongs to the security scheme swallows every
// annotation that follows it, so tags declared after it vanish from the
// generated document without a word of warning.
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
// The version above is the version of the API contract, not the version of a
// release: Release Please tags the repository and never touches this file. Bump
// it by hand when a change breaks a consumer.
//
// The host, basePath and schemes annotations are deprecated in OpenAPI 3.1 and
// are replaced by the servers entries above. Do not open a prose line with an
// at-sign: the swaggo formatter reads it as an annotation and indents it into
// the block above. The two local servers are separate because the two functions
// listen on separate ports during development; once deployed they share one API
// Gateway domain.
//
// No @license is declared: the licence is a decision of the project built from
// this template, and claiming one here would be a claim nobody made.
