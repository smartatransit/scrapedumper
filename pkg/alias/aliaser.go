package alias

import (
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	cache "github.com/patrickmn/go-cache"
	"github.com/sahilm/fuzzy"
)

type Alias struct {
	NamedElementID uint `gorm:"not null"`
	Alias          string
}
type Aliases []Alias

func (a Aliases) String(i int) string {
	return a[i].Alias
}
func (a Aliases) Len() int {
	return len(a)
}

type ID struct {
	ID uint `json:"-" gorm:"primary_key"`
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

type AliasLookup interface {
	FindNamedElementByRoughName(kind, name string) (uint, error)
}

func (a AliasLookupAgent) FindNamedElementByRoughName(kind, name string) (uint, error) {
	var aliases Aliases
	if aliasesI, found := a.cache.Get("kind"); found {
		aliases = aliasesI.(Aliases)
	} else {
		a.db.Find(&aliases, "named_element_type = ?", kind)
		a.cache.Set("kind", aliases, cache.DefaultExpiration)
	}

	matches := fuzzy.FindFrom(name, aliases)
	if len(matches) > 0 {
		var result Alias
		a.db.Table("aliases").Find(&result, aliases[matches[0].Index].NamedElementID)

		return result.NamedElementID, nil
	} else {
		err := errors.New(fmt.Sprintf("No station match for name %s", name))
		return 0, err
	}
}
