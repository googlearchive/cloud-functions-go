## Locating `worker.js` in the Google Cloud Function environment

First, prepare an `index.js` that is valid, but has a runtime error:

``` js
exports.helloWorld = function helloWorld (req, res) {
    console.log(abc) // `abc` is not defined
}
```

Next, deploy it, e.g. with `gcloud beta functions deploy`

Finally, trigger it and look at your Google Cloud Function log. You'll see a stacktrace of your error, for example:

```
ReferenceError: abc is not defined at helloWorld (/user_code/index.js:20:14)
  at /var/tmp/worker/worker.js:635:7
  at /var/tmp/worker/worker.js:619:9
  at _combinedTickCallback (internal/process/next_tick.js:73:7)
  at process._tickDomainCallback (internal/process/next_tick.js:128:9)
```

From that stacktrace, see that in our example the wrapper running your Google Cloud Function is located at `/var/tmp/worker/worker.js`.

## Contents of `worker.js`

Although `fs.readFileSync` is not available in the given node js environment (i.e. `var fs = require("fs")` will fail) but we can use the Go runtime to give us the file contents:

``` go
func main() {
	flag.Parse()
	http.HandleFunc(nodego.HTTPTrigger, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		f, err := os.Open("/var/tmp/worker/worker.js")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		_, err = io.Copy(w, f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	nodego.TakeOver()
}
```

If you trigger your Google Cloud Function via http, you will get the contents of `/var/tmp/worker/worker.js` in the http response body.
