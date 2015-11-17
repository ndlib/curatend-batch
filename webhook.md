Batch Ingest Webhook RFC

This document proposes a way to add webhooks to the batch ingest process, so
another application can receive notification of the current state of batch
ingest jobs.

1. An file named `WEBHOOK` consisting of a list of URLs separated by newlines
may be present inside a batch ingest job's directory. If the file is present
and readable, each URL will be called in sequence whenever the state changes
for that job.

2. The URIs may use HTTP Basic auth. If so then use the form:
https://user:password@example.com/whatever?q=blah The URIs may have query
parameters. If a url does not start with "http" or "https", the string
"http://" is prefixed.

3. The (current) states of a job are "queue", "processing", "error", and
"success". The hooks will be called on entry to "processing", "error", and
"success". More states may be added in the future.

4. If an error occurs sending a webhook, then that particular call will be
lost. But, webhooks for any remaining URLs in the webhook file will still be
called and the job will still change state. In particular, THE RECEIPT OF A
WEBHOOK IS NOT GUARANTEED, and a URL will receive at most ONE notice of any
given state change.

5. The URL will receive a HTTP POST message with a message body with content
type "application/json". The JSON payload will be a JSON hash with the
following fields at a minimum:

    host        The hostname of the machine running the batch job processor. e.g. "libvirt8".
    version     The version of the batch job processor. String.
    job_name    The name of the job having the state change. String.
    job_state   The new state of the job. String.
    job_log     The job log. Only present if the job_state is "error". String.

6. Example JSON message

    {
        "host" : "libvirt8.library.nd.edu",
        "version" : "1.0.1",
        "job_name" : "ingest-45",
        "job_state" : "processing"
    }

6. Note the CSRF must be disabled on the rails side, since no token is in the
response.

7. Is anything missing?
