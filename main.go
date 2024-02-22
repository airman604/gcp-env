package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

var SM_PREFIX = "sm://"

var _secretManagerClient *secretmanager.Client

func getSecretManagerClient(ctx context.Context) *secretmanager.Client {
	if _secretManagerClient == nil {
		var err error
		_secretManagerClient, err = secretmanager.NewClient(ctx)
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}
		// defer client.Close()
	}

	return _secretManagerClient
}

func doSecretsManagerSubstitution(ctx context.Context, input string) (string, error) {

	log.Printf("INFO: doing variable substitution for %v\n", input)

	input = strings.TrimPrefix(input, SM_PREFIX)
	secret, value, parseJSON := strings.Cut(input, "#")
	if !strings.Contains(secret, "/versions/") {
		// pick latest version of the secret by default
		secret = secret + "/versions/latest"
	}
	log.Printf("INFO: parsed secret name is %v", secret)

	client := getSecretManagerClient(ctx)

	// Build the request.
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secret,
	}

	// Call the API.
	secretData, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret %q: %v", secret, err)
	}

	secretValue := secretData.Payload.Data
	result := string(secretValue)
	if parseJSON {
		log.Printf("INFO: JSON parsing is enabled, extracting field: %v", value)

		parsedSecret := make(map[string]string)
		err := json.Unmarshal(secretValue, &parsedSecret)
		if err != nil {
			return "", fmt.Errorf("failed to parse secret data: %v", err)
		}

		var ok bool
		result, ok = parsedSecret[value]
		if !ok {
			return "", fmt.Errorf("failed to parse secret data: key %q not present in JSON", value)
		}
	}

	return result, nil
}

func main() {

	ctx := context.Background()
	substitutions := make(map[string]string)
	for _, e := range os.Environ() {
		_split := strings.SplitN(e, "=", 2)
		e_name, e_val := _split[0], _split[1]
		if strings.HasPrefix(e_val, SM_PREFIX) {
			var err error
			e_val, err = doSecretsManagerSubstitution(ctx, e_val)
			if err != nil {
				log.Fatalf("ERROR: %v\n", err)
			}
			substitutions[e_name] = e_val
		} else {
			substitutions[e_name] = e_val
		}
	}

	for e_name, e_val := range substitutions {
		err := os.Setenv(e_name, e_val)
		if err != nil {
			log.Fatalf("ERROR: %v\n", err)
		}
	}

	progExec := os.Args[1]
	progArgs := os.Args[1:]
	progPath, err := exec.LookPath(progExec)
	if err != nil {
		log.Fatalf("ERROR: %v\n", err)
	}

	err = syscall.Exec(progPath, progArgs, os.Environ())
	if err != nil {
		log.Fatalf("ERROR: %v\n", err)
	}
}
