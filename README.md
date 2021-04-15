# op - Simple REST client for 1Password

# Usage

    # create client
    client, _ := op.NewClient("http://localhost:8080", op.WithTokenFromEnv())

    # get vault
    vault, _ := c.Vault.Get("UIO")

    # search for items
    items, _ := vault.Item.Get(FilterByTags("root_token"))

    # create password
    _ = vault.Item.Password.Create(
            "MySecretPass", "s3cr3tp@ss",
            WithTags("tags", "are", "cool"),
    )

    # create note
    _ = vault.Item.Note.Create(
            "MyNote", "could be any string, perhaps JSON :)",
            WithTags("tags", "are", "cool"),
    )

# Useful Resources
* https://support.1password.com/connect-api-reference/
* https://github.com/1Password/connect-sdk-go
* https://support.1password.com/secrets-automation/
* https://support.1password.com/connect-deploy-docker
* https://github.com/1Password/vault-plugin-secrets-onepassword

