package v1alpha2

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/klog"
	"net/http"
	"strings"
)

type jenkinsProxy struct {
	client       core.JenkinsCore
	host         string
	scheme       string
	roundTripper http.RoundTripper
}

func newJenkinsProxy(client core.JenkinsCore, host, scheme string, roundTripper http.RoundTripper) *jenkinsProxy {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}
	return &jenkinsProxy{
		client:       client,
		host:         host,
		scheme:       scheme,
		roundTripper: roundTripper,
	}
}

func (p *jenkinsProxy) proxyWithDevOps(request *restful.Request, response *restful.Response) {
	u := request.Request.URL
	devopsPath := request.PathParameter("devops")
	u.Host = p.host
	u.Scheme = p.scheme
	u.Path = strings.Replace(request.Request.URL.Path, fmt.Sprintf("/kapis/%s/%s/devops/%s/jenkins",
		GroupVersion.Group, GroupVersion.Version, devopsPath), "", 1)
	u.Path = strings.Replace(u.Path, fmt.Sprintf("/%s/devops/%s/jenkins",
		GroupVersion.Version, devopsPath), "", 1)
	httpProxy := proxy.NewUpgradeAwareHandler(u, p.roundTripper, false, false, &errorResponder{})

	if err := p.client.AuthHandle(request.Request); err != nil {
		msg := "failed to set auth header for Jenkins API request"
		klog.V(4).Infof("%s, error: %v", msg, err)
		_, _ = response.Write([]byte(msg))
		return
	}
	httpProxy.ServeHTTP(response, request.Request)
}
