package op

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/go-resty/resty/v2"
)

const (
	TokenKEY = "OPASSWORD_ACCESS_TOKEN"
	VERSION = "/v1"
)

// ItemCategory Represents the template of the Item
type ItemCategory string

const (
	Password             ItemCategory = "PASSWORD"
	SecureNote           ItemCategory = "SECURE_NOTE"
	Custom               ItemCategory = "CUSTOM"
)

type Client struct {
	conn *resty.Client
	endpoint   string
	token      string

	Vault *VaultClient
}


type ItemClient struct {
	client *Client
}

type VaultClient struct {
	client  *Client
	currentID string
	currentName string

	Item *ItemClient
}

type ClientOption func(*Client)

func WithTokenFromEnv() ClientOption {
	return func(c *Client) {
		value := os.Getenv(TokenKEY)
		c.token = value
	}
}

func WithToken(value string) ClientOption {
	return func(c *Client) {
		c.token = value
	}
}

func NewClient(endpoint string, options ...ClientOption) (*Client, error) {
	client := resty.New()
	c := &Client{
		conn: client,
		endpoint: endpoint + VERSION,
	}
	item := &ItemClient{client: c}
	vault := &VaultClient{
		client: c,
		Item: item,
	}
	c.Vault = vault

	for _, option := range options {
		option(c)
	}
	c.conn.SetAuthToken(c.token).SetHeader("Accept", "application/json")
	return c, nil
}

func (c *VaultClient) Get(name string) (*VaultClient, error) {
	var vaults []Vault
	resp, err := c.client.conn.R().SetResult(&vaults).Get(c.client.endpoint + "/vaults")
	if err != nil {
		return nil, errors.Wrapf(err, "connection error")
	}
	if resp.IsError() {
		return nil, errors.Wrapf(err, "http error")
	}
	for _, v := range vaults {
		if name == v.Name() {
			c.currentName = name
			c.currentID = v.ID()
			return c, nil
		}
	}
	return nil, errors.New("no results")
}

func (c *VaultClient) UUID() string {
	return c.currentID
}

func (c *VaultClient) Name() string {
	return c.currentName
}

type Vault map[string]interface{}

func (v Vault) Name() string {
	name, _ := v["name"].(string)
	return name
}

func (v Vault) ID() string {
	id, _ := v["id"].(string)
	return id
}

type Item map[string]interface{}

func (i Item) Tags() []string {
	// FIXME: Convert to []string directly if possible
	t, ok := i["tags"].([]interface{})
	if !ok {
		return []string{}
	}

	tags := make([]string, len(t))

	for n, tag := range t {
		tags[n] = tag.(string)
	}
	return tags
}

func (i Item) Title() string {
	title, _ := i["title"].(string)
	return title
}

func (i Item) Category() ItemCategory {
	category, _ := i["category"].(ItemCategory)
	return category
}

func (i Item) ID() string {
	id, _ := i["id"].(string)
	return id
}

type Filter func([]Item) []Item

func stringExists(slice []string, val string) bool {
	for _, str := range slice {
		if str == val {
			return true
		}
	}
	return false
}

func itemContainsTag(item Item, tag string) bool {
	return stringExists(item.Tags(), tag)
}

func FilterByTags(tags ...string) Filter {
	return func(items []Item) []Item {
		newItems := make([]Item, 0)
	next:
		for _, i := range items {
			for _, tag := range tags {
				if !itemContainsTag(i, tag) {
					continue next
				}
			}
			newItems = append(newItems, i)
		}
		return newItems
	}
}

func FilterByTitle(name string) Filter {
	return func(items []Item) []Item {
		newItems := make([]Item, 0)
		for _, i := range items {
			if i.Title() == name {
				newItems = append(newItems, i)
			}
		}
		return newItems
	}
}

func FilterByCategory(name ItemCategory) Filter {
	return func(items []Item) []Item {
		newItems := make([]Item, 0)
		for _, i := range items {
			if i.Category() == name {
				newItems = append(newItems, i)
			}
		}
		return newItems
	}
}

func (c *ItemClient) GetDetails(uuid string) (Item, error) {
	var item Item
	vaultID := c.client.Vault.UUID()
	path := fmt.Sprintf("/vaults/%s/items/%s", vaultID, uuid)
	resp, err := c.client.conn.R().SetResult(&item).Get(c.client.endpoint + path)
	if err != nil {
		return nil, errors.Wrapf(err, "connection error")
	}
	if resp.IsError() {
		return nil, errors.Wrapf(err, "http error")
	}
	return item, nil
}

func (c *ItemClient) Get(filters ...Filter) ([]Item, error) {
	var items []Item
	uuid := c.client.Vault.UUID()
	path := fmt.Sprintf("/vaults/%s/items", uuid)
	resp, err := c.client.conn.R().SetResult(&items).Get(c.client.endpoint + path)
	if err != nil {
		return nil, errors.Wrapf(err, "connection error")
	}
	if resp.IsError() {
		return nil, errors.Wrapf(err, "http error")
	}
	itemsDetailed := make([]Item, len(items))
	for n, i := range items {
		item, err := c.GetDetails(i.ID())
		if err != nil {
			return nil, err
		}
		itemsDetailed[n] = item
	}
	filtered := applyFilters(itemsDetailed, filters...)
	return filtered, nil
}

func applyFilters(items []Item, filters ...Filter) []Item {
	if len(filters) == 0 {
		return items
	}
	filter := filters[0]
	if len(filters) == 1 {
		return filter(items)
	}
	items = filter(items)
	rest := filters[1:]
	return applyFilters(items, rest...)
}