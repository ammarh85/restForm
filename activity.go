package rest

import (
	"bytes"
	"crypto/tls"
	//"crypto/x509"
	"encoding/json"
	"io"
	"io/ioutil"
	//"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

// log is the default package logger
var log = logger.GetLogger("activity-tibco-rest")

const (
	methodGET    = "GET"
	methodPOST   = "POST"
	methodPUT    = "PUT"
	methodPATCH  = "PATCH"
	methodDELETE = "DELETE"

	ivMethod      = "method"
	ivURI         = "uri"
	ivPathParams  = "pathParams"
	ivQueryParams = "queryParams"
	ivContent     = "content"
	ivFormInput   = "formInput"
	ivParams      = "params"

	ovResult = "result"
	ovStatusCode = "statusCode"
)

var validMethods = []string{methodGET, methodPOST, methodPUT, methodPATCH, methodDELETE}

// RESTActivity is an Activity that is used to invoke a REST Operation
// inputs : {method,uri,params}
// outputs: {result}
type RESTActivity struct {
	metadata *activity.Metadata
}

// NewActivity creates a new RESTActivity
func NewActivity(metadata *activity.Metadata) activity.Activity {
	return &RESTActivity{metadata: metadata}
}

// Metadata returns the activity's metadata
func (a *RESTActivity) Metadata() *activity.Metadata {
	return a.metadata
}

// Eval implements api.Activity.Eval - Invokes a REST Operation
func (a *RESTActivity) Eval(context activity.Context) (done bool, err error) {

	method := strings.ToUpper(context.GetInput(ivMethod).(string))
	uri := context.GetInput(ivURI).(string)

	containsParam := strings.Index(uri, "/:") > -1


	if containsParam {

		val := context.GetInput(ivPathParams)

		if val == nil {
			val = context.GetInput(ivParams)

			if val == nil {
				err := activity.NewError("Path Params not specified, required for URI: " + uri, "", nil)
				return false, err
			}
		}

		pathParams := val.(map[string]string)
		uri = BuildURI(uri, pathParams)
	}

	if context.GetInput(ivQueryParams) != nil {
		queryParams := context.GetInput(ivQueryParams).(map[string]string)

		qp := url.Values{}

		for key, value := range queryParams {
			qp.Set(key, value)
		}

		uri = uri + "?" + qp.Encode()
	}

	log.Debugf("REST Call: [%s] %s\n", method, uri)

	var reqBody io.Reader

	contentType := "application/json; charset=UTF-8"

	if method == methodPOST || method == methodPUT || method == methodPATCH {

		content := context.GetInput(ivContent)

		contentType = getContentType(content)

		if content != nil {
			if str, ok := content.(string); ok {
				reqBody = bytes.NewBuffer([]byte(str))
			} else {
				b, _ := json.Marshal(content) //todo handle error
				reqBody = bytes.NewBuffer([]byte(b))
			}
		} else {

			formInfo := context.GetInput(ivFormInput)
			if formInfo != nil {
				if str, ok := formInfo.(string); ok {
					reqBody = bytes.NewBuffer([]byte(str))
					contentType = "application/x-www-form-urlencoded"
				} else {
					b, _ := json.Marshal(formInfo) //todo handle error
					reqBody = bytes.NewBuffer([]byte(b))
				}

			}
		}
	} else {
		reqBody = nil
	}

	req, err := http.NewRequest(method, uri, reqBody)

	if reqBody != nil {
		req.Header.Set("Content-Type", contentType)
	}

	client := buildClient()
	//resp, err := client.Get("https://localhost:8080/hello")
	//client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Info("ERROR BHOW!!!!")
		panic(err)
	}
	defer resp.Body.Close()

	log.Debug("response Status:", resp.Status)
	statusArray := strings.Fields(resp.Status)
	status, err := strconv.Atoi(statusArray[0])
	if err != nil {
		// handle error
		err := activity.NewError("Unknow Status Code Returned " + resp.Status, "", nil)
		return false, err
	}

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result interface{}

	d := json.NewDecoder(bytes.NewReader(respBody))
	d.UseNumber()
	err = d.Decode(&result)

	//json.Unmarshal(respBody, &result)

	log.Debug("response Body:", result)

	context.SetOutput(ovResult, result)
	context.SetOutput(ovStatusCode, status)

	return true, nil
}

////////////////////////////////////////////////////////////////////////////////////////
// Utils

//todo just make contentType a setting
func getContentType(replyData interface{}) string {

	contentType := "application/json; charset=UTF-8"

	switch v := replyData.(type) {
	case string:
		if !strings.HasPrefix(v, "{") && !strings.HasPrefix(v, "[") {
			contentType = "text/plain; charset=UTF-8"
		}
	case int, int64, float64, bool, json.Number :
		contentType = "text/plain; charset=UTF-8"
	default:
		contentType = "application/json; charset=UTF-8"
	}

	return contentType
}

func methodIsValid(method string) bool {

	if !stringInList(method, validMethods) {
		return false
	}

	//validate path

	return true
}

func stringInList(str string, list []string) bool {
	for _, value := range list {
		if value == str {
			return true
		}
	}
	return false
}

// BuildURI is a temporary crude URI builder
func BuildURI(uri string, values map[string]string) string {

	var buffer bytes.Buffer
	buffer.Grow(len(uri))

	addrStart := strings.Index(uri, "://")

	i := addrStart + 3

	for i < len(uri) {
		if uri[i] == '/' {
			break
		}
		i++
	}

	buffer.WriteString(uri[0:i])

	for i < len(uri) {
		if uri[i] == ':' {
			j := i + 1
			for j < len(uri) && uri[j] != '/' {
				j++
			}

			if i+1 == j {

				buffer.WriteByte(uri[i])
				i++
			} else {

				param := uri[i+1: j]
				value := values[param]
				buffer.WriteString(value)
				if j < len(uri) {
					buffer.WriteString("/")
				}
				i = j + 1
			}

		} else {
			buffer.WriteByte(uri[i])
			i++
		}
	}

	return buffer.String()
}

func buildClient() *http.Client {
	// Set up our own certificate pool
	log.Info("MAKE BHOW CLIENT!!!!")
	//tlsConfig := &tls.Config{RootCAs: x509.NewCertPool()}
	//transport := &http.Transport{TLSClientConfig: tlsConfig}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	return client
}
