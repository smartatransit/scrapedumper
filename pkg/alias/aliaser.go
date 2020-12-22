package alias

import (
	"fmt"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/sahilm/fuzzy"
	"gorm.io/gorm"
)

type Alias struct {
	NamedElementType string `gorm:"not null"`
	NamedElementID   uint   `gorm:"not null"`
	Alias            string
}
type Aliases []Alias

func (a Aliases) String(i int) string {
	return normalize(a[i].Alias)
}
func (a Aliases) Len() int {
	return len(a)
}

type AliasLookupAgent struct {
	db    *gorm.DB
	cache *cache.Cache
}

func New(
	db *gorm.DB,
) AliasLookupAgent {
	db = db.Set("gorm:auto_preload", true)

	return AliasLookupAgent{
		db: db,

		// Refetch aliases from the database daily
		cache: cache.New(24*time.Hour, 24*time.Hour),
	}
}

func normalize(s string) string {
	return strings.TrimSpace(strings.TrimSuffix(strings.ToUpper(s), "STATION"))
}

type AliasLookup interface {
	FindNamedElementByRoughName(kind, name string) (*uint, error)
}

func (a AliasLookupAgent) getAliases(kind string) (Aliases, error) {
	var aliases map[string]Aliases
	if aliasesI, found := a.cache.Get(""); found {
		aliases = aliasesI.(map[string]Aliases)
	} else {
		var aliasList Aliases
		if err := a.db.Find(&aliasList).Error; err != nil {
			return nil, fmt.Errorf("Failed fetching aliases for type %s: %w", kind, err)
		}

		aliases = map[string]Aliases{}
		for _, alias := range aliasList {
			aliases[alias.NamedElementType] = append(aliases[alias.NamedElementType], alias)
		}

		a.cache.Set("", aliases, cache.DefaultExpiration)
	}

	return aliases[kind], nil
}

func (a AliasLookupAgent) FindNamedElementByRoughName(kind, name string) (*uint, error) {
	aliases, err := a.getAliases(kind)
	if err != nil {
		return nil, err
	}

	matches := fuzzy.FindFrom(normalize(name), aliases)
	if len(matches) > 0 {
		return &aliases[matches[0].Index].NamedElementID, nil
	}

	return nil, fmt.Errorf("No %s match for name %s", kind, name)
}
