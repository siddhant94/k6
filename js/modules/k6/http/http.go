/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2016 Load Impact
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package http

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"reflect"

	"fmt"
	"net/http/httputil"

	"github.com/loadimpact/k6/js/common"
	log "github.com/sirupsen/logrus"
)

var (
	typeString                     = reflect.TypeOf("")
	typeURL                        = reflect.TypeOf(URL{})
	typeMapKeyStringValueInterface = reflect.TypeOf(map[string]interface{}{})
)

const (
	HTTP_METHOD_GET                    = "GET"
	HTTP_METHOD_POST                   = "POST"
	HTTP_METHOD_PUT                    = "PUT"
	HTTP_METHOD_DELETE                 = "DELETE"
	HTTP_METHOD_HEAD                   = "HEAD"
	HTTP_METHOD_PATCH                  = "PATCH"
	HTTP_METHOD_OPTIONS                = "OPTIONS"
	OCSP_STATUS_GOOD                   = "good"
	OCSP_STATUS_REVOKED                = "revoked"
	OCSP_STATUS_SERVER_FAILED          = "server_failed"
	OCSP_STATUS_UNKNOWN                = "unknown"
	OCSP_REASON_UNSPECIFIED            = "unspecified"
	OCSP_REASON_KEY_COMPROMISE         = "key_compromise"
	OCSP_REASON_CA_COMPROMISE          = "ca_compromise"
	OCSP_REASON_AFFILIATION_CHANGED    = "affiliation_changed"
	OCSP_REASON_SUPERSEDED             = "superseded"
	OCSP_REASON_CESSATION_OF_OPERATION = "cessation_of_operation"
	OCSP_REASON_CERTIFICATE_HOLD       = "certificate_hold"
	OCSP_REASON_REMOVE_FROM_CRL        = "remove_from_crl"
	OCSP_REASON_PRIVILEGE_WITHDRAWN    = "privilege_withdrawn"
	OCSP_REASON_AA_COMPROMISE          = "aa_compromise"
	SSL_3_0                            = "ssl3.0"
	TLS_1_0                            = "tls1.0"
	TLS_1_1                            = "tls1.1"
	TLS_1_2                            = "tls1.2"
)

type HTTPCookie struct {
	Name, Value, Domain, Path string
	HttpOnly, Secure          bool
	MaxAge                    int
	Expires                   int64
}

type HTTPRequestCookie struct {
	Name, Value string
	Replace     bool
}

type HTTP struct {
	SSL_3_0                            string `js:"SSL_3_0"`
	TLS_1_0                            string `js:"TLS_1_0"`
	TLS_1_1                            string `js:"TLS_1_1"`
	TLS_1_2                            string `js:"TLS_1_2"`
	OCSP_STATUS_GOOD                   string `js:"OCSP_STATUS_GOOD"`
	OCSP_STATUS_REVOKED                string `js:"OCSP_STATUS_REVOKED"`
	OCSP_STATUS_SERVER_FAILED          string `js:"OCSP_STATUS_SERVER_FAILED"`
	OCSP_STATUS_UNKNOWN                string `js:"OCSP_STATUS_UNKNOWN"`
	OCSP_REASON_UNSPECIFIED            string `js:"OCSP_REASON_UNSPECIFIED"`
	OCSP_REASON_KEY_COMPROMISE         string `js:"OCSP_REASON_KEY_COMPROMISE"`
	OCSP_REASON_CA_COMPROMISE          string `js:"OCSP_REASON_CA_COMPROMISE"`
	OCSP_REASON_AFFILIATION_CHANGED    string `js:"OCSP_REASON_AFFILIATION_CHANGED"`
	OCSP_REASON_SUPERSEDED             string `js:"OCSP_REASON_SUPERSEDED"`
	OCSP_REASON_CESSATION_OF_OPERATION string `js:"OCSP_REASON_CESSATION_OF_OPERATION"`
	OCSP_REASON_CERTIFICATE_HOLD       string `js:"OCSP_REASON_CERTIFICATE_HOLD"`
	OCSP_REASON_REMOVE_FROM_CRL        string `js:"OCSP_REASON_REMOVE_FROM_CRL"`
	OCSP_REASON_PRIVILEGE_WITHDRAWN    string `js:"OCSP_REASON_PRIVILEGE_WITHDRAWN"`
	OCSP_REASON_AA_COMPROMISE          string `js:"OCSP_REASON_AA_COMPROMISE"`

	isMultipart bool
}

func New() *HTTP {
	return &HTTP{
		SSL_3_0:                            SSL_3_0,
		TLS_1_0:                            TLS_1_0,
		TLS_1_1:                            TLS_1_1,
		TLS_1_2:                            TLS_1_2,
		OCSP_STATUS_GOOD:                   OCSP_STATUS_GOOD,
		OCSP_STATUS_REVOKED:                OCSP_STATUS_REVOKED,
		OCSP_STATUS_SERVER_FAILED:          OCSP_STATUS_SERVER_FAILED,
		OCSP_STATUS_UNKNOWN:                OCSP_STATUS_UNKNOWN,
		OCSP_REASON_UNSPECIFIED:            OCSP_REASON_UNSPECIFIED,
		OCSP_REASON_KEY_COMPROMISE:         OCSP_REASON_KEY_COMPROMISE,
		OCSP_REASON_CA_COMPROMISE:          OCSP_REASON_CA_COMPROMISE,
		OCSP_REASON_AFFILIATION_CHANGED:    OCSP_REASON_AFFILIATION_CHANGED,
		OCSP_REASON_SUPERSEDED:             OCSP_REASON_SUPERSEDED,
		OCSP_REASON_CESSATION_OF_OPERATION: OCSP_REASON_CESSATION_OF_OPERATION,
		OCSP_REASON_CERTIFICATE_HOLD:       OCSP_REASON_CERTIFICATE_HOLD,
		OCSP_REASON_REMOVE_FROM_CRL:        OCSP_REASON_REMOVE_FROM_CRL,
		OCSP_REASON_PRIVILEGE_WITHDRAWN:    OCSP_REASON_PRIVILEGE_WITHDRAWN,
		OCSP_REASON_AA_COMPROMISE:          OCSP_REASON_AA_COMPROMISE,
	}
}

func (*HTTP) XCookieJar(ctx *context.Context) *HTTPCookieJar {
	return newCookieJar(ctx)
}

func (*HTTP) CookieJar(ctx context.Context) *HTTPCookieJar {
	state := common.GetState(ctx)
	return &HTTPCookieJar{state.CookieJar, &ctx}
}

func (*HTTP) mergeCookies(req *http.Request, jar *cookiejar.Jar, reqCookies map[string]*HTTPRequestCookie) map[string][]*HTTPRequestCookie {
	allCookies := make(map[string][]*HTTPRequestCookie)
	for _, c := range jar.Cookies(req.URL) {
		allCookies[c.Name] = append(allCookies[c.Name], &HTTPRequestCookie{Name: c.Name, Value: c.Value})
	}
	for key, reqCookie := range reqCookies {
		if jc := allCookies[key]; jc != nil && reqCookie.Replace {
			allCookies[key] = []*HTTPRequestCookie{{Name: key, Value: reqCookie.Value}}
		} else {
			allCookies[key] = append(allCookies[key], &HTTPRequestCookie{Name: key, Value: reqCookie.Value})
		}
	}
	return allCookies
}

func (*HTTP) setRequestCookies(req *http.Request, reqCookies map[string][]*HTTPRequestCookie) {
	for _, cookies := range reqCookies {
		for _, c := range cookies {
			req.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value})
		}
	}
}

func (*HTTP) debugRequest(state *common.State, req *http.Request, description string) {
	if state.Options.HttpDebug.String != "" {
		dump, err := httputil.DumpRequestOut(req, state.Options.HttpDebug.String == "full")
		if err != nil {
			log.Fatal(err)
		}
		logDump(description, dump)
	}
}

func (*HTTP) debugResponse(state *common.State, res *http.Response, description string) {
	if state.Options.HttpDebug.String != "" && res != nil {
		dump, err := httputil.DumpResponse(res, state.Options.HttpDebug.String == "full")
		if err != nil {
			log.Fatal(err)
		}
		logDump(description, dump)
	}
}

func logDump(description string, dump []byte) {
	fmt.Printf("%s:\n%s\n", description, dump)
}
