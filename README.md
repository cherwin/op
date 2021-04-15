# op - Simple REST client for 1Password

# Usage

```golang
package main

import (
	"github.com/cherwin/op"
)

func main() {
	// Create Client instance
	client, _ := op.NewClient("http://localhost:8080", op.WithTokenFromEnv())

	// Get Vault instance
	vault, _ := client.Vault.Get("TestVault")

	// Search for items
	_, _ = vault.Item.Get(op.FilterByTags("some", "tag"))
	
	// Create password
	_, _ = vault.Item.Password.Create(
		"MySecretPass", "s3cr3tp@ss",
		op.WithTags("always", "use", "tags"),
	)

	// Create note
	_, _ = vault.Item.Note.Create(
		"MyNote", "could be any string, perhaps JSON :)",
		op.WithTags("they", "are", "useful"),
	)
}
```

# Useful Resources
* https://support.1password.com/connect-api-reference/
* https://github.com/1Password/connect-sdk-go
* https://support.1password.com/secrets-automation/
* https://support.1password.com/connect-deploy-docker
* https://github.com/1Password/vault-plugin-secrets-onepassword

