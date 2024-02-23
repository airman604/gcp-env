# gcp-env

This is a small utility to help populate environment variables using secrets stored
in Google Secret Manager. Inspired by similar tools for AWS:
[telia-oss/aws-env](https://github.com/telia-oss/aws-env),
[remind101/ssm-env](https://github.com/remind101/ssm-env), 
[sendgrid/aws-env](https://github.com/sendgrid/aws-env).

## How it works

**gcp-env** will loop through the environment and replace any variables prefixed
with `sm://` with their secret value from Secret Manager. Google [Application
Default Credentials](https://cloud.google.com/docs/authentication/provide-credentials-adc)
have to be configured and have access to the referenced secrets.

Secrets can be referenced using either full version resource ID, i.e.
`projects/PROJECT_ID/secrets/SECRET_NAME/versions/VERSION_ID`, or just
`projects/PROJECT_ID/secrets/SECRET_NAME`, in which case **gcp-env** will retrieve
"latest" version of the secret.

If the environment variable ends in `#some-key`, then the value of the secret
will be parsed as JSON and the value of the referenced key (`some-key`) from JSON
will be substituted.

## Usage

1. Grab **gcp-env** binary for your platform from the
[releases page](https://github.com/airman604/gcp-env/releases). Place it in a
directory that's in the PATH and make it executable.

2. Start your application with **gcp-env**:

```bash
export MY_SECRET=sm://projects/1234567890123/secrets/my-secret
export MY_JSON_SECRET="sm://projects/1234567890123/secrets/my-secret#key"
gcp-env COMMAND [PARAMETERS]
```

This will populate all the secrets in the environment, and start specified
`COMMAND` with the provided `PARAMETERS`. The populated secrets are only made
available to the command and do not persist once the process exits.

## More examples

Run multiple commands with one secrets lookup:

```bash
gcp-env sh -c "command1; command2; command3"
```

Substitute references to secrets in a configuration file:

```bash
cat input.cfg.template | gcp-env envsubst > output.cfg
```

Store whole configuration file in Secret Manager:

```bash
export MY_SECRET_CONFIG=sm://...
# using single quotes is important here!
gcp-env sh -c 'echo "$MY_SECRET_CONFIG" > secret.cfg'
```


Inject secrets into a container:

```bash
export MY_SECRET=sm://projects/1234567890123/secrets/my-secret
gcp-env docker run -d -e MY_SECRET alpine
```

Use gcp-env within the container:

```Dockerfile
FROM alpine

RUN wget https://github.com/airman604/gcp-env/releases/download/v0.0.3/gcp-env_0.0.3_linux_amd64 -O /usr/local/bin/gcp-env && \
  chmod +x /usr/local/bin/gcp-env

ENTRYPOINT ["gcp-env"]
```