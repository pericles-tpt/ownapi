package node

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pericles-tpt/ownapi/secrets"
	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pkg/errors"
)

var (
	supportedMethods       = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodHead}
	supportedMethodsString = strings.Join(supportedMethods, ", ")
)

type HttpNodeConfig struct {
	BaseNodeProps

	RawBaseUrl      string    `json:"raw_url"`
	RawUrlPathParts *[]string `json:"raw_url_path_parts"`
	Method          string    `json:"method"`
	// TODO: Currently headers and queryParams are constant, but it's better for them to be variable
	//			e.g. ?location=Sydney or ?location=Melbourne for a weather node
	Headers                  *map[string]string `json:"headers"`
	Params                   *map[string]string `json:"query_params"`
	UserAgent                *string            `json:"user_agent"`
	UseHeadAndCacheResponses bool               `json:"use_head_and_cache_responses"`

	OutputKey string `json:"output_key"`
}

type HTTPNode struct {
	Config     HttpNodeConfig `json:"config"`
	CacheCheck *HeadProps     `json:"cache_check"`
}

type HeadProps struct {
	LastModified       *time.Time `json:"last_modified"`
	LastModifiedAsEtag *string    `json:"last_modified_as_etag"`

	Etag string `json:"etag"`
}

type HTTPStatus struct {
	Code    int
	Message string
}

func CreateHTTPNode(propMap map[string]any, cfg HttpNodeConfig) (HTTPNode, error) {
	var ret HTTPNode

	cfg, err := utility.OverrideTypeFromJSONMap(cfg, propMap)
	if err != nil {
		return ret, errors.Wrap(err, "failed to override `cfg` with map values")
	}
	ret.Config = cfg

	if ret.Config.OutputKey == "" {
		return ret, errors.New("invalid `output_key` provided, must be non-empty")
	}
	if !strings.HasPrefix(ret.Config.OutputKey, "output:") {
		ret.Config.OutputKey = fmt.Sprintf("output:%s", ret.Config.OutputKey)
	}

	_, err = url.Parse(ret.Config.RawBaseUrl)
	if err != nil {
		return ret, errors.Wrap(err, "invalid url provided")
	}

	_, contains := utility.Contains(ret.Config.Method, supportedMethods)
	if !contains {
		return ret, fmt.Errorf("invalid method '%s' provided, not one of: %s", ret.Config.Method, supportedMethodsString)
	}

	if ret.Config.UseHeadAndCacheResponses {
		ret.CacheCheck = &HeadProps{}
	}

	err = ret.regenerateHash()
	if err != nil {
		return ret, errors.Wrap(err, "failed to generate hash for new `HttpNodeConfig`")
	}

	return ret, nil
}

func (hn *HTTPNode) Trigger(propMap map[string]any) (map[string]any, error) {
	var outputMap = map[string]any{}
	// TODO: Should this return `isCachedOutput` as the first return value?
	// var isCachedOutput bool
	newCfg, err := utility.OverrideTypeFromJSONMap(hn.Config, propMap)
	if err != nil {
		return outputMap, errors.Wrap(err, "failed to override `cfg` with map values")
	}
	defer func(hn *HTTPNode, oldCfg HttpNodeConfig) {
		hn.Config = oldCfg
	}(hn, hn.Config)
	hn.Config = newCfg

	if hn.Config.UseHeadAndCacheResponses {
		modified, err := hn.triggerHead(propMap)
		if !modified && err == nil {
			// Failing to read from local cache is NOT a failure condition, can still do full request
			lastResult := hn.readCachedResponseData()
			if lastResult != nil {
				propMap[hn.Config.OutputKey] = *lastResult
				return outputMap, nil
			}
		}
	}

	urlWithParams, err := hn.maybeAddURLPathAndQueryParams(propMap)
	if err != nil {
		return outputMap, errors.Wrap(err, "failed to modify base url")
	}
	req, err := http.NewRequest(hn.Config.Method, urlWithParams, nil)
	if err != nil {
		return outputMap, err
	}

	if hn.Config.Headers != nil {
		for k, v := range *hn.Config.Headers {
			req.Header.Set(k, v)
		}
	}

	if hn.Config.UserAgent != nil {
		req.Header.Set("User-Agent", *hn.Config.UserAgent)
	}

	client := &http.Client{}
	resp, err := replaceSecretsThenDo(client, req)
	if err != nil {
		return outputMap, err
	}
	defer resp.Body.Close()

	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !isSuccess {
		return outputMap, fmt.Errorf("response code for HTTP request not in 'success' range - code: %d, message: '%s'", resp.StatusCode, resp.Status)
	}

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return outputMap, err
	}

	if hn.Config.UseHeadAndCacheResponses {
		hn.writeCachedResponseData(buf.Bytes())
	}

	outputMap[hn.Config.OutputKey] = buf.Bytes()
	fmt.Println("RESP SIZE IS: ", buf.Len())

	return outputMap, nil
}

func (hn *HTTPNode) triggerNoCache(propMap map[string]any) (map[string]any, error) {
	var outputMap = map[string]any{}
	newCfg, err := utility.OverrideTypeFromJSONMap(hn.Config, propMap)
	if err != nil {
		return outputMap, errors.Wrap(err, "failed to override `cfg` with map values")
	}
	defer func(hn *HTTPNode, oldCfg HttpNodeConfig) {
		hn.Config = oldCfg
	}(hn, hn.Config)
	hn.Config = newCfg

	urlWithParams, err := hn.maybeAddURLPathAndQueryParams(propMap)
	if err != nil {
		return outputMap, errors.Wrap(err, "failed to modify base url")
	}
	req, err := http.NewRequest(hn.Config.Method, urlWithParams, nil)
	if err != nil {
		return outputMap, err
	}

	if hn.Config.Headers != nil {
		for k, v := range *hn.Config.Headers {
			req.Header.Set(k, v)
		}
	}

	if hn.Config.UserAgent != nil {
		req.Header.Set("User-Agent", *hn.Config.UserAgent)
	}

	client := &http.Client{}
	resp, err := replaceSecretsThenDo(client, req)
	if err != nil {
		return outputMap, err
	}
	defer resp.Body.Close()

	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !isSuccess {
		return outputMap, fmt.Errorf("response code for HTTP request not in 'success' range - code: %d, message: '%s'", resp.StatusCode, resp.Status)
	}

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return outputMap, err
	}

	outputMap[hn.Config.OutputKey] = buf.Bytes()
	fmt.Println("RESP SIZE IS: ", buf.Len())
	return outputMap, nil
}

func (hn *HTTPNode) triggerHead(propMap map[string]any) (bool, error) {
	var modified bool

	urlWithParams, err := hn.maybeAddURLPathAndQueryParams(propMap)
	if err != nil {
		return modified, errors.Wrap(err, "failed to modify base url")
	}
	req, err := http.NewRequest(http.MethodHead, urlWithParams, nil)
	if err != nil {
		return modified, err
	}

	if hn.Config.Headers != nil {
		for k, v := range *hn.Config.Headers {
			req.Header.Set(k, v)
		}
	}

	if hn.Config.UserAgent != nil {
		req.Header.Set("User-Agent", *hn.Config.UserAgent)
	}

	client := &http.Client{}
	resp, err := replaceSecretsThenDo(client, req)
	if err != nil {
		return modified, err
	}
	defer resp.Body.Close()

	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !isSuccess {
		return modified, fmt.Errorf("response code for HTTP request not in 'success' range - code: %d, message: '%s'", resp.StatusCode, resp.Status)
	}

	modified, err = hn.updateHeadPropsAndGetModified(resp)
	if err != nil {
		return modified, err
	}

	if modified {
		err = hn.regenerateHash()
		if err != nil {
			return modified, errors.Wrap(err, "failed to generate new hash after `HeadProps` was modified")
		}
		// TODO: If `hn.CacheCheck` is modified it should be flushed to disk
	}

	return modified, nil
}

func (hn *HTTPNode) regenerateHash() error {
	// Remove cache file for old file
	if hn.Config.Hash != "" {
		cachedFilePath := fmt.Sprintf("%s/%s", httpResponseCacheOutputPath, hn.Config.Hash)
		err := os.Remove(cachedFilePath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	copyForHash := HTTPNode{}
	copyForHash = *hn
	copyForHash.Config.Hash = ""
	copyForHash.CacheCheck = nil

	nodeBytes, err := json.Marshal(copyForHash)
	if err != nil {
		return errors.Wrap(err, "failed to marshal node to bytes")
	}

	hash := sha1.New()
	_, err = hash.Write(nodeBytes)
	if err != nil {
		return errors.Wrap(err, "failed to write bytes to hash")
	}
	newHashBytes := hash.Sum(nil)

	hn.Config.Hash = fmt.Sprintf("%x", newHashBytes)
	return nil
}

func (hn *HTTPNode) maybeAddURLPathAndQueryParams(propMap map[string]any) (string, error) {
	generatedUrl, err := hn.maybeGetRawUrlPath(propMap)
	if err != nil {
		return generatedUrl, err
	}
	if hn.Config.Params == nil {
		return generatedUrl, nil
	}

	paramStrings := make([]string, 0, len(*hn.Config.Params))
	for k, partOrKey := range *hn.Config.Params {
		var val any = partOrKey
		if strings.HasPrefix(partOrKey, "input:") {
			var found bool
			if val, found = propMap[partOrKey]; !found {
				return hn.Config.RawBaseUrl, fmt.Errorf("invalid key '%s' in `Params`, not found in `propMap`", partOrKey)
			}
		}
		paramStrings = append(paramStrings, fmt.Sprintf("%s=%v", k, val))
	}
	return fmt.Sprintf("%s?%s", generatedUrl, strings.Join(paramStrings, "&")), nil
}

func (hn *HTTPNode) maybeGetRawUrlPath(propMap map[string]any) (string, error) {
	if hn.Config.RawUrlPathParts == nil {
		return hn.Config.RawBaseUrl, nil
	}
	evalParts := make([]string, 0, len(*hn.Config.RawUrlPathParts))
	for _, partOrKey := range *hn.Config.RawUrlPathParts {
		var val any = partOrKey
		if strings.HasPrefix(partOrKey, "input:") {
			var found bool
			if val, found = propMap[partOrKey]; !found {
				return hn.Config.RawBaseUrl, fmt.Errorf("invalid key '%s' in `RawUrlPathParts`, not found in `propMap`", partOrKey)
			}
		}
		evalParts = append(evalParts, fmt.Sprint(val))
	}
	return fmt.Sprintf("%s/%s", hn.Config.RawBaseUrl, strings.Join(evalParts, "/")), nil
}

func (hn *HTTPNode) updateHeadPropsAndGetModified(resp *http.Response) (bool, error) {
	// Check for ETag and Last-Modified
	var (
		modified bool

		maybeEtag         = resp.Header.Get("ETag")
		maybeLastModified = resp.Header.Get("Last-Modified")
	)
	if maybeEtag == "" && maybeLastModified == "" {
		return modified, errors.New("response missing both required properties for a HEAD request in the header: ETag and Last-Modified")
	}
	var (
		etagChanged, lastModifiedChanged, lastModifiedAsEtagChanged bool
	)
	etagChanged = hn.CacheCheck.Etag != maybeEtag
	hn.CacheCheck.Etag = maybeEtag
	if maybeLastModified != "" {
		lastModified, err := time.Parse(time.RFC1123, maybeLastModified)
		invalidLastModifiedFormat := err != nil // -> Treat 'Last-Modified' as another ETag
		if invalidLastModifiedFormat {
			lastModifiedAsEtagChanged = hn.CacheCheck.LastModifiedAsEtag == nil || *hn.CacheCheck.LastModifiedAsEtag != maybeLastModified
			hn.CacheCheck.LastModified = nil
			hn.CacheCheck.LastModifiedAsEtag = &maybeLastModified
		} else {
			lastModifiedChanged = hn.CacheCheck.LastModified == nil || hn.CacheCheck.LastModified.Before(lastModified)
			hn.CacheCheck.LastModified = &lastModified
			hn.CacheCheck.LastModifiedAsEtag = nil
		}
	}
	modified = modified || etagChanged || lastModifiedAsEtagChanged || lastModifiedChanged

	return modified, nil
}

func (hn *HTTPNode) readCachedResponseData() *[]byte {
	cachedFilePath := fmt.Sprintf("%s/%s", httpResponseCacheOutputPath, hn.Config.Hash)
	data, err := os.ReadFile(cachedFilePath)
	if err != nil {
		return nil
	}
	return &data
}

func (hn *HTTPNode) writeCachedResponseData(data []byte) {
	cachedFilePath := fmt.Sprintf("%s/%s", httpResponseCacheOutputPath, hn.Config.Hash)
	os.Remove(cachedFilePath)

	err := os.WriteFile(cachedFilePath, data, 0660)
	if err != nil {
		fmt.Println("Failed to write file: ", err)
	}
}

func replaceSecretsThenDo(client *http.Client, req *http.Request) (*http.Response, error) {
	// TODO: Replace secrets in Form, PostForm, MultipartForm
	changed, maybeNewUrl, err := secrets.MaybeReplaceSecretsInString(req.URL.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to replace secrets in url")
	}
	if changed {
		req.URL, err = url.Parse(maybeNewUrl)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse URL with replaced secrets")
		}
	}

	for k, vl := range req.Header {
		for i, v := range vl {
			_, newHeaderValue, err := secrets.MaybeReplaceSecretsInString(v)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to replace secrets in header with key '%s'", k)
			}
			vl[i] = newHeaderValue
		}

		if len(vl) > 0 {
			req.Header.Del(k)
			req.Header.Set(k, vl[0])
			for i := 1; i < len(vl); i++ {
				req.Header.Add(k, vl[i])
			}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (hn *HTTPNode) Changed(propsMap map[string]any) bool {
	return true
}
func (hn *HTTPNode) revert(changed *bool, propsMap map[string]any) {
}
func (hn *HTTPNode) GetTrigger() *Trigger {
	return hn.Config.NodeTrigger
}
