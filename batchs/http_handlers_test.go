package batchs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

type testInfo struct {
	URL    string
	Body   string
	Status int
}

type testHeadInfo struct {
	URL    string
	Status int
}

// Throw up a canned webserver and file system to test
// the http calls, then run the tests

var testFS string
var testServer *httptest.Server

func TestMain(m *testing.M) {

	testFS, _ = ioutil.TempDir("", "test-batchs")
	defer os.RemoveAll(testFS)

	fileQ := NewFileQueue(testFS)
	err := fileQ.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Port is there to satisfy interface- won't be used
	server := RESTServer{
		QueuePath:  fileQ,
		PortNumber: "15000",
		Version:    "testing",
	}

	// start httptest server
	testServer = httptest.NewServer(server.addRoutes())

	// Run tests
	ret := m.Run()

	// clean up
	testServer.Close()

	os.Exit(ret)
}

func TestGetJobs(t *testing.T) {

	// Get routes to test, and expected bodys

	fileContent := []byte("this is content for Gets, baby")

	getTests := []testInfo{
		{"/", "CurateND Batch (testing)\n", 200},
		{"/jobs", "[\"queuedjob\",\"testjob1\"]\n", 200},
		{"/jobs/testjob1", "{\"Name\":\"testjob1\",\"Status\":\"success\"}\n", 200},
		{"/jobs/testjob1/files/testfile1", string(fileContent), 200},
		{"/jobs/testjob1/files/testfile2", "", 404},
		{"/jobs/testjob1/files/", "[\"testfile1\"]\n", 200},
		{"/jobs/testjob1/files/something/../..", "[\"testfile1\"]\n", 200},
		{"/jobs/non-existent/", "Not Found\n", 404},
		{"/jobs/non-existent/files/", "", 404},
		{"/jobs/queuedjob", "{\"Name\":\"queuedjob\",\"Status\":\"queue\"}\n", 200},
		{"/jobs/queuedjob/", "{\"Name\":\"queuedjob\",\"Status\":\"queue\"}\n", 200},
		{"/jobs/queuedjob/files/testfile1", "Cannot access queued jobs", 409},
	}

	// test setup
	createJobFile(t, testFS, "success", "testjob1", "testfile1", fileContent)
	createJobFile(t, testFS, "queue", "queuedjob", "testfile1", fileContent)

	for _, thisTest := range getTests {
		t.Log("Testing GET ", thisTest.URL)
		testBody := getbody(t, "GET", thisTest.URL, thisTest.Status)

		if testBody != thisTest.Body {
			t.Errorf("Received %#v, expected %#v", testBody, thisTest.Body)
		}
	}
}

func TestHeadJobs(t *testing.T) {

	// Get routes to test, and expected bodys

	fileContent := []byte("this is content for Heads, baby")

	headTests := []testHeadInfo{
		{"/", 200},
		{"/jobs", 405},
		{"/jobs/testjob1", 200},
		{"/jobs/testjob1/files/testfile1", 200},
		{"/jobs/testjob1/files/testfile2", 404},
		{"/jobs/testjob2/files/mydir", 200},
		{"/jobs/testjob2/files/mydir/testfile2", 200},
	}

	// test setup
	createJobFile(t, testFS, "success", "testjob1", "testfile1", fileContent)
	createJobFile(t, testFS, "queue", "queuedjob", "testfile1", fileContent)
	// A little cheating here, I need to test directory behavior
	createJobFile(t, testFS, "success", "testjob2/mydir", "testfile2", fileContent)

	for _, thisTest := range headTests {
		t.Log("Testing HEAD ", thisTest.URL)
		checkStatus(t, "HEAD", thisTest.URL, thisTest.Status)
	}
}

func TestPutJobs(t *testing.T) {
	t.Log("Testing PUT /jobs/testjob2 (new)")
	checkStatus(t, "PUT", "/jobs/testjob2", 200)

	t.Log("Testing PUT /jobs/testjob2 (exists)")
	checkStatus(t, "PUT", "/jobs/testjob2", 403)

	t.Log("Testing PUT /jobs/testjob2/dir1/file1")
	uploadstring(t, "PUT", "/jobs/testjob2/files/dir1/file1", "ph'nglui mglw'nafh Cthulhu R'lyeh wgah'nagl fhtagn", 200)
}

func TestDeleteJobs(t *testing.T) {
	t.Log("Testing DELETE /jobs/testjob3")

	fileContent := []byte("this is content for , baby")

	// test setup
	createJobFile(t, testFS, "data", "testjob3", "testfile1", fileContent)

	t.Log("Testing DELETE /jobs/testjob3/files/testfile1")
	checkStatus(t, "GET", "/jobs/testjob3/files/testfile1", 200)
	checkStatus(t, "DELETE", "/jobs/testjob3/files/testfile1", 200)
	checkStatus(t, "DELETE", "/jobs/testjob3/files/testfile2", 200)
	checkStatus(t, "GET", "/jobs/testjob3/files/testfile1", 404)
	checkStatus(t, "DELETE", "/jobs/testjob3/files/testfile1", 200)

	t.Log("Testing DELETE /jobs/testjob3")
	checkStatus(t, "GET", "/jobs/testjob3", 200)
	checkStatus(t, "DELETE", "/jobs/testjob3", 200)
	checkStatus(t, "GET", "/jobs/testjob3", 404)
}

func TestSubmitHandler(t *testing.T) {
	t.Log("Testing POST /jobs/testjob4")

	fileContent := []byte("this is content, baby")

	// test setup
	createJobFile(t, testFS, "data", "testjob4", "testfile1", fileContent)

	checkStatus(t, "POST", "/jobs/testjob4/queue", 200)
	checkStatus(t, "POST", "/jobs/testjob4/queue", 409)
	checkStatus(t, "POST", "/jobs/testjob5/queue", 404)
}

// some test utility functions

func createJobFile(t *testing.T, testFS, dir, id, fileName string, fileContent []byte) {

	err := os.MkdirAll(path.Join(testFS, dir, id), 0744)
	if err != nil {
		t.Fatalf("Could not create directory")
	}

	err = ioutil.WriteFile(path.Join(testFS, dir, id, fileName), fileContent, 0755)
	if err != nil {
		t.Fatalf("Could not create file")
	}
}

func checkRoute(t *testing.T, verb, route string, expstatus int) *http.Response {
	req, err := http.NewRequest(verb, testServer.URL+route, nil)
	if err != nil {
		t.Fatal("Problem creating request", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(route, err)
		return nil
	}
	if resp.StatusCode != expstatus {
		t.Errorf("%s: Expected status %d and received %d",
			route,
			expstatus,
			resp.StatusCode)
		resp.Body.Close()
		return nil
	}
	return resp
}

func getbody(t *testing.T, verb, route string, expstatus int) string {
	resp := checkRoute(t, verb, route, expstatus)
	if resp != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(route, err)
		}
		resp.Body.Close()
		return string(body)
	}
	return ""
}

func checkStatus(t *testing.T, verb, route string, expstatus int) {
	resp := checkRoute(t, verb, route, expstatus)
	if resp != nil {
		resp.Body.Close()
	}
}

func uploadstring(t *testing.T, verb, route, s string, statuscode int) {

	req, err := http.NewRequest(verb, testServer.URL+route, strings.NewReader(s))
	if err != nil {
		t.Fatal("Problem creating request", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(route, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != statuscode {
		t.Errorf("%s: Received status %d, expected %d",
			route,
			resp.StatusCode,
			statuscode)
	}
}
