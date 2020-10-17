package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type svcCheck struct {
	Subsets []struct {
		Addresses []struct {
			IP string
		}
	}
}

type podsList struct {
	Items []struct {
		Status struct {
			PodIP string
		}
	}
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func authorized(clientIP string) bool {
	ns := os.Getenv("NAMESPACE")
	query := "?labelSelector=kubernetes-api-access%3Denabled"
	url := fmt.Sprintf("/api/v1/namespaces/%s/pods%s", ns, query)
	headers := make(map[string]string)
	var body []byte
	code, resp := handleK8sReq("GET", url, body, headers)
	if code != 200 {
		return false
	}
	pods := podsList{}
	err := json.Unmarshal([]byte(resp), &pods)
	if err != nil {
		log.Println("Unmarshal err:", err.Error())
	}

	for _, pod := range pods.Items {
		if clientIP == pod.Status.PodIP {
			return true
		}
	}
	log.Println("NOT authorized for IP", clientIP)
	return false
}

func health(w http.ResponseWriter, r *http.Request) {
	log.Println("client ip", strings.Split(getClientIP(r), ":")[0], "uri", r.URL.RequestURI(), "status", 200)
	w.WriteHeader(200)
	return
}

func waitTillInService() {
	log.Println("Waiting till out of service")
	for healthCheckStatus() {
		log.Println("Still in service")
	}
	log.Println("Done with wait")
}

func healthCheckStatus() bool {
	ns := os.Getenv("NAMESPACE")
	svc := os.Getenv("KUBERNETES_SERVICE_NAME")
	ip := os.Getenv("POD_IP")
	succ := 0
	attempts := 5
	for succ < 2 && attempts > 0 {
		attempts--
		if inService(ns, svc, ip) == 200 {
			succ++
		} else {
			succ = 0
		}
		time.Sleep(1 * time.Second)
	}
	if succ == 2 {
		return true
	}
	return false
}

func serviceCheck(w http.ResponseWriter, r *http.Request) {
	pathParams := strings.Split(r.URL.RequestURI(), "/")
	if len(pathParams) != 5 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ns := pathParams[2]
	svc := pathParams[3]
	ip := pathParams[4]
	st := inService(ns, svc, ip)
	log.Println("client ip", getClientIP(r), "uri", r.URL.RequestURI(), "status", st)
	w.WriteHeader(inService(ns, svc, ip))
}

func inService(ns string, svc string, ip string) (code int) {
	check := svcCheck{}
	headers := make(map[string]string)
	var body []byte
	url := fmt.Sprintf("/api/v1/namespaces/%s/endpoints/%s", ns, svc)
	code, resp := handleK8sReq("GET", url, body, headers)
	if code != 200 {
		return
	}
	json.Unmarshal([]byte(resp), &check)
	if check.Subsets == nil {
		return 404
	}
	for _, addr := range check.Subsets[0].Addresses {
		log.Println(addr.IP)
		if addr.IP == ip {
			return 200
		}
	}
	return 404
}

func k8sEndpoints(w http.ResponseWriter, r *http.Request) {
	if !authorized(strings.Split(getClientIP(r), ":")[0]) {
		log.Println("client ip", getClientIP(r), "uri", r.URL.RequestURI(), "Auth status", "false", "status", 401)
		w.WriteHeader(401)
		return
	}
	headers := make(map[string]string)
	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			headers[name] = value
		}
	}
	body, _ := ioutil.ReadAll(r.Body)
	code, resp := handleK8sReq(r.Method, r.URL.RequestURI(), body, headers)
	log.Println(code, resp)
	w.WriteHeader(code)
	fmt.Fprintf(w, resp)
	log.Println("client ip", getClientIP(r), "uri", r.URL.RequestURI(), "status", code)
}

func handleK8sReq(method string, url string, payload []byte, headers map[string]string) (statusCode int, body string) {
	pemPath := "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	certs := x509.NewCertPool()

	pemData, err := ioutil.ReadFile(pemPath)
	if err != nil {
		// do error
	}
	certs.AppendCertsFromPEM(pemData)
	mTLSConfig := &tls.Config{
		RootCAs: certs,
	}

	tr := &http.Transport{TLSClientConfig: mTLSConfig}
	client := &http.Client{Transport: tr}

	authToken, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	headers["Authorization"] = fmt.Sprintf("Bearer %s", authToken)
	url = "https://kubernetes.default.svc.cluster.local" + url
	req, _ := http.NewRequest(method, url, bytes.NewBuffer((payload)))
	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	resp, err := client.Do(req)
	if err != nil {

	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func main() {
	http.HandleFunc("/health", health)
	http.HandleFunc("/inservice/", serviceCheck)
	http.HandleFunc("/", k8sEndpoints)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("Server Started")

	<-done
	log.Println("Stopping server")
	waitTillInService()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("Server Exited Properly")
}
