package playground

import (
	"embed"
	"html/template"
	"net/http"
	"net/url"
)

//go:embed statics/*
var StaticFiles embed.FS

// Handler 返回静态文件服务器
func StaticHandler() http.Handler {
	return http.FileServer(http.FS(StaticFiles))
}

var page = template.Must(template.New("graphiql").Parse(`<!DOCTYPE html>
<html>
  <head>
  	<meta charset="utf-8">
  	<title>{{.title}}</title>
	<style>
		body {
			height: 100%;
			margin: 0;
			width: 100%;
			overflow: hidden;
		}

		#graphiql {
			height: 100vh;
		}
	</style>
	<script
		src="statics/react.production.min.js"
		crossorigin="anonymous"
	></script>
	<script
		src="/statics/react-dom.production.min.js"
		crossorigin="anonymous"
	></script>
    <link
		rel="stylesheet"
		href="/statics/graphiql.min.css"
		crossorigin="anonymous"
	/>
  </head>
  <body>
    <div id="graphiql">Loading...</div>

	<script
		src="/statics/graphiql.min.js"
		crossorigin="anonymous"
	></script>

    <script>
{{- if .endpointIsAbsolute}}
      const url = {{.endpoint}};
      const subscriptionUrl = {{.subscriptionEndpoint}};
{{- else}}
      const url = location.protocol + '//' + location.host + {{.endpoint}};
      const wsProto = location.protocol == 'https:' ? 'wss:' : 'ws:';
      const subscriptionUrl = wsProto + '//' + location.host + {{.endpoint}};
{{- end}}
{{- if .fetcherHeaders}}
      const fetcherHeaders = {{.fetcherHeaders}};
{{- else}}
      const fetcherHeaders = undefined;
{{- end}}
{{- if .uiHeaders}}
      const uiHeaders = {{.uiHeaders}};
{{- else}}
      const uiHeaders = undefined;
{{- end}}

      const fetcher = GraphiQL.createFetcher({ url, subscriptionUrl, headers: fetcherHeaders });
      ReactDOM.render(
        React.createElement(GraphiQL, {
          fetcher: fetcher,
          isHeadersEditorEnabled: true,
          shouldPersistHeaders: true,
		  headers: JSON.stringify(uiHeaders, null, 2)
        }),
        document.getElementById('graphiql'),
      );
    </script>
  </body>
</html>
`))

// Handler responsible for setting up the playground
func Handler(title, endpoint string) http.HandlerFunc {
	return HandlerWithHeaders(title, endpoint, nil, nil)
}

// HandlerWithHeaders sets up the playground.
// fetcherHeaders are used by the playground's fetcher instance and will not be visible in the UI.
// uiHeaders are default headers that will show up in the UI headers editor.
func HandlerWithHeaders(
	title, endpoint string,
	fetcherHeaders, uiHeaders map[string]string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html; charset=UTF-8")
		err := page.Execute(w, map[string]any{
			"title":                title,
			"endpoint":             endpoint,
			"fetcherHeaders":       fetcherHeaders,
			"uiHeaders":            uiHeaders,
			"endpointIsAbsolute":   endpointHasScheme(endpoint),
			"subscriptionEndpoint": getSubscriptionEndpoint(endpoint),
			"version":              "3.7.0",
			"cssSRI":               "sha256-Dbkv2LUWis+0H4Z+IzxLBxM2ka1J133lSjqqtSu49o8=",
			"jsSRI":                "sha256-qsScAZytFdTAEOM8REpljROHu8DvdvxXBK7xhoq5XD0=",
			"reactSRI":             "sha256-S0lp+k7zWUMk2ixteM6HZvu8L9Eh//OVrt+ZfbCpmgY=",
			"reactDOMSRI":          "sha256-IXWO0ITNDjfnNXIu5POVfqlgYoop36bDzhodR6LW5Pc=",
		})
		if err != nil {
			panic(err)
		}
	}
}

// endpointHasScheme checks if the endpoint has a scheme.
func endpointHasScheme(endpoint string) bool {
	u, err := url.Parse(endpoint)
	return err == nil && u.Scheme != ""
}

// getSubscriptionEndpoint returns the subscription endpoint for the given
// endpoint if it is parsable as a URL, or an empty string.
func getSubscriptionEndpoint(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}

	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	default:
		u.Scheme = "ws"
	}

	return u.String()
}
