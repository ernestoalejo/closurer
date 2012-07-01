package main

import (
	"fmt"
	"path/filepath"
)

func TestHandler(r *Request) error {
	r.W.Header().Set("Content-Type", "text/html; charset=utf-8")

	name := r.Req.URL.Path[6:]
	name = name[:len(name)-5] + ".js"

	data := map[string]interface{}{
		"Name": name,
	}
	return r.ExecuteTemplate(TEST_TEMPLATE, data)
}

func TestAllHandler(r *Request) error {
	r.W.Header().Set("Content-Type", "text/html; charset=utf-8")

	tests, err := ScanTests()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"AllTests": tests,
	}
	return r.ExecuteTemplate(GLOBAL_TEST_TEMPLATE, data)
}

// Search for "_test.js" files and relativize them to
// the root directory. It replaces the .js ext with .html.
func ScanTests() ([]string, error) {
	tests, err := Scan(conf.RootJs, "_test.js")
	if err != nil {
		return nil, err
	}

	for i, test := range tests {
		// Relativize the path adding .html instead of .js
		p, err := filepath.Rel(conf.RootJs, test[:len(test)-2]+"html")
		if err != nil {
			return nil, fmt.Errorf("cannot relativize %s: %s", test, err)
		}
		tests[i] = p
	}

	return tests, nil
}

func TestListHandler(r *Request) error {
	r.W.Header().Set("Content-Type", "text/html; charset=utf-8")

	tests, err := ScanTests()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"AllTests": tests,
	}
	return r.ExecuteTemplate(TEST_LIST_TEMPLATE, data)
}

const TEST_TEMPLATE = `
{{define "base"}}
<!DOCTYPE html>
<html>
<head>

	<meta charset="utf-8">
	<title>Unit Test</title>

	<script type="text/javascript" src="/input/base.js"></script>
	<script type="text/javascript" src="/input/{{.Name}}"></script>

</head>
<body>

</body>
</html>
{{end}}
`

const GLOBAL_TEST_TEMPLATE = `
{{define "base"}}
<!DOCTYPE html>
<html>
<head>

	<meta charset="utf-8">
	<title>Test Runner</title>

	<style type="text/css">

		.goog-testrunner {
		  background-color: #EEE;
		  border: 1px solid #999;
		  padding: 10px;
		  padding-bottom: 25px;
		}

		.goog-testrunner-progress {
		  width: auto;
		  height: 20px;
		  background-color: #FFF;
		  border: 1px solid #999;
		}

		.goog-testrunner-progress table {
		  width: 100%;
		  height: 20px;
		  border-collapse: collapse;
		}

		.goog-testrunner-buttons {
		  margin-top: 7px;
		}

		.goog-testrunner-buttons button {
		  width: 75px;
		}

		.goog-testrunner-log,
		.goog-testrunner-report,
		.goog-testrunner-stats {
		  margin-top: 7px;
		  width: auto;
		  height: 400px;
		  background-color: #FFF;
		  border: 1px solid #999;
		  font: normal medium monospace;
		  padding: 5px;
		  overflow: auto;  /* Opera doesn't support overflow-y. */
		  overflow-y: scroll;
		  overflow-x: auto;
		}

		.goog-testrunner-report div {
		  margin-bottom: 6px;
		  border-bottom: 1px solid #999;
		}

		.goog-testrunner-stats table {
		  margin-top: 20px;
		  border-collapse: collapse;
		  border: 1px solid #EEE;
		}

		.goog-testrunner-stats td,
		.goog-testrunner-stats th {
		  padding: 2px 6px;
		  border: 1px solid #F0F0F0;
		}

		.goog-testrunner-stats th {
		  font-weight: bold;
		}

		.goog-testrunner-stats .center {
		  text-align: center;
		}

		.goog-testrunner-progress-summary {
		  font: bold small sans-serif;
		}

		.goog-testrunner iframe {
		  position: absolute;
		  left: -640px;
		  top: -480px;
		  width: 640px;
		  height: 480px;
		  margin: 0;
		  border: 0;
		  padding: 0;
		}

		.goog-testrunner-report-failure {
		  color: #900;
		}

		.goog-testrunner-reporttab,
		.goog-testrunner-logtab,
		.goog-testrunner-statstab {
		  float: left;
		  width: 50px;
		  height: 16px;
		  text-align: center;
		  font: normal small arial, helvetica, sans-serif;
		  color: #666;
		  background-color: #DDD;
		  border: 1px solid #999;
		  border-top: 0;
		  cursor: pointer;
		}

		.goog-testrunner-reporttab,
		.goog-testrunner-logtab {
		  border-right: 0;
		}

		.goog-testrunner-activetab {
		  font-weight: bold;
		  color: #000;
		  background-color: #CCC;
		}

		h1 {
		  font: normal x-large arial, helvetica, sans-serif;
		  margin: 0;
		}

		p, form {
		  font: normal small sans-serif;
		  margin: 0;
		}

		#header {
		  position: absolute;
		  right: 10px;
		  top: 13px;
		  color: #090;
		}

		#footer {
		  margin-top: 8px;
		}

		.warning {
		  font-size: 14px;
		  font-weight: bold;
		  width: 80%;
		}

	</style>

	<script type="text/javascript" src="/input/base.js"></script>
	<script type="text/javascript">
		goog.require('goog.userAgent.product');
		goog.require('goog.testing.MultiTestRunner');
	</script>

</head>
<body>

	<h1>All JsUnit Tests</h1>
	<p id="header">
		<a href="/list">List of Individual Tests</a> |
		<a href="/">Home</a>
	</p>

	<div id="runner"></div>

	<form id="footer" onsubmit="return false;">

		Settings:<br>

		<input type="checkbox" name="hidepasses" id="hidepasses">
		<label for="hidepasses">Hide passes</label><br>

		<input type="checkbox" name="parallel" id="parallel" checked>
		<label for="parallel">Run in parallel</label>
		<small>(timing stats not available if enabled)</small><br>

		<input type="text" name="filter" id="filter" value="">
		<label for="filter">Run only tests for path</label>

	</form>

	<script type="text/javascript">

		(function() {
			var allTests = {{.AllTests}};

		  var hidePassesInput = document.getElementById('hidepasses');
		  var parallelInput = document.getElementById('parallel');
		  var filterInput = document.getElementById('filter');
		  
		  function setFilterFunction() {
		    var matchValue = filterInput.value || '';
		    testRunner.setFilterFunction(function(testPath) {
		      return testPath.indexOf(matchValue) > -1;
		    });
		  }
		  
		  // Create a test runner and render it.
		  var testRunner = new goog.testing.MultiTestRunner()
		      .setName(document.title)
		      .setPoolSize(parallelInput.checked ? 8 : 1)
		      .setStatsBucketSizes(5, 500)
		      .setHidePasses(hidePassesInput.checked)
		      .addTests(allTests);
		  testRunner.render(document.getElementById('runner'));
		  
		  goog.events.listen(hidePassesInput, 'click', function(e) {
		    testRunner.setHidePasses(e.target.checked);
		  });
		  
		  goog.events.listen(parallelInput, 'click', function(e) {
		    testRunner.setPoolSize(e.target.checked ? 8 : 1);
		  });
		  
		  goog.events.listen(filterInput, 'keyup', setFilterFunction);
		  setFilterFunction();
		})();

	</script>

</body>
</html>
{{end}}
`
const TEST_LIST_TEMPLATE = `
{{define "base"}}
<!DOCTYPE html>
<html>
<head>

	<meta charset="utf-8">
	<title>Unit Tests List</title>

</head>
<body>

	<h1>All JsUnit Tests</h1>
	
	<ul>
		{{range .AllTests}}
			<li><a href="/test/{{.}}">{{.}}</a></li>
		{{end}}
	</ul>

</body>
</html>
{{end}}
`
