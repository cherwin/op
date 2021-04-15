package op

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVaultClient_Get(t *testing.T) {
	c, err := NewClient("http://localhost:8080", WithTokenFromEnv())
	require.Nil(t, err)
	vault, err := c.Vault.Get("UIO")
	require.Nil(t, err)
	fmt.Println(vault.Name(), vault.UUID())
}

func TestItemClient_Get(t *testing.T) {
	c, err := NewClient("http://localhost:8080", WithTokenFromEnv())
	require.Nil(t, err)
	vault, err := c.Vault.Get("UIO")
	require.Nil(t, err)
	fmt.Println(vault.Name(), vault.UUID())
	items, err := c.Vault.Item.Get(
		FilterByTags("root_token", "titanium"),
	)
	require.Nil(t, err)
	for _, item := range items {
		fmt.Println("*", item)
	}
}

func TestNoteClient_Create(t *testing.T) {
	c, err := NewClient("http://localhost:8080", WithTokenFromEnv())
	require.Nil(t, err)
	vault, err := c.Vault.Get("UIO")
	require.Nil(t, err)
	item, err := vault.Item.Note.Create("foo", "some test", WithTags("foo", "bar"))
	require.Nil(t, err)
	fmt.Println(item)
}

func TestPasswordClient_Create(t *testing.T) {
	c, err := NewClient("http://localhost:8080", WithTokenFromEnv())
	require.Nil(t, err)
	vault, err := c.Vault.Get("UIO")
	require.Nil(t, err)
	item, err := vault.Item.Password.Create("new pass", "swordfish", WithTags("travolta"))
	require.Nil(t, err)
	fmt.Println(item)
}

func TestItemContainsTag(t *testing.T) {
	tags := []string{"foo", "bar"}
	item := Item{"tags": tags}
	found := itemContainsTag(item, "foo")
	assert.True(t, found)

	found = itemContainsTag(item, "bar")
	assert.True(t, found)

	found = itemContainsTag(item, "baz")
	assert.False(t, found)
}

func TestNewClientWithBadToken(t *testing.T) {
	c, err := NewClient("http://localhost:8080", WithToken("foo"))
	require.Nil(t, err)
	_, err = c.Vault.Get("Foo")
	require.NotNil(t, err)
	assert.EqualError(t, err, "http error: 401")
}

func TestFilterByTags(t *testing.T) {
	items := []Item{
		{
			"tags": []string{
				"foo", "bar",
			},
		},
		{
			"tags": []string{
				"bar", "quux",
			},
		},
		{
			"title": "Foo",
			"tags": []string{
				"spam",
			},
			"category": SecureNote,
		},
	}
	items = applyFilters(items,
		FilterByTags("spam"),
		FilterByTitle("Foo"),
		FilterByCategory(SecureNote),
	)
	fmt.Println(items)
}
