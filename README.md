# op - Simple REST client for 1Password

# Usage

    client, _ := op.NewClient("http://localhost:8080", op.WithTokenFromEnv())
    vault, _ := c.Vault.Get("UIO")
    items, _ := vault.Item.Get(FilterByTags("root_token"))

# Useful Resources
* https://support.1password.com/connect-api-reference/
* https://github.com/1Password/connect-sdk-go
* https://support.1password.com/secrets-automation/
* https://support.1password.com/connect-deploy-docker
* https://github.com/1Password/vault-plugin-secrets-onepassword

