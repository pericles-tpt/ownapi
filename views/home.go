package views

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/pericles-tpt/ownapi/config"
	"github.com/pericles-tpt/rterror"
)

func Home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	frontendUrlParts := strings.Split(config.GetStaticFrontendURL(), "/")
	frontendRoutingPrefix := fmt.Sprintf("/%s", frontendUrlParts[len(frontendUrlParts)-1])
	if len(frontendUrlParts) == 1 {
		frontendRoutingPrefix = ""
	}

	var (
		backendUrl         = config.GetBackendURL()
		backendUrlProtocol = "https://"
	)
	if config.GetIsDev() {
		backendUrl = config.GetLocalBackendURL()
		backendUrlProtocol = "http://"
	}

	pageData := map[string]interface{}{
		"Title":                 "Demo",
		"BackendURL":            backendUrlProtocol + backendUrl,
		"FrontendRoutingPrefix": frontendRoutingPrefix,
	}

	tmpl, err := template.New("index.html").Funcs(template.FuncMap{}).ParseFiles("./src/index.html")
	if err != nil {
		rterror.PrintPrependErrorWithRuntimeInfo(err, "failed to parse template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, pageData)
	if err != nil {
		rterror.PrintPrependErrorWithRuntimeInfo(err, "failed to execute template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
