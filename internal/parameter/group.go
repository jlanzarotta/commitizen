package parameter

import (
	standarderrs "errors"

	"github.com/charmbracelet/huh"
	"github.com/shipengqi/golib/strutil"

	"github.com/shipengqi/commitizen/internal/errors"
	"github.com/shipengqi/commitizen/internal/helpers"
)

type Group struct {
	Name      string    `yaml:"name"       json:"name"       mapstructure:"name"`
	DependsOn DependsOn `yaml:"depends_on" json:"depends_on" mapstructure:"depends_on"`
}

func NewGroup(name string) *Group {
	return &Group{Name: name}
}

func (g *Group) Validate() []error {
	var errs []error
	if strutil.IsEmpty(g.Name) {
		errs = append(errs, standarderrs.New("the group missing required field: name"))
	}
	for _, v := range g.DependsOn.OrConditions {
		errs = append(errs, v.Validate()...)
	}
	for _, v := range g.DependsOn.AndConditions {
		errs = append(errs, v.Validate()...)
	}
	return errs
}

func (g *Group) Render(all map[FieldKey]huh.Field, fields []huh.Field) *huh.Group {
	for _, v := range g.DependsOn.OrConditions {
		v.fields = all
	}
	for _, v := range g.DependsOn.AndConditions {
		v.fields = all
	}

	group := huh.NewGroup(fields...)
	group.WithHideFunc(func() bool {
		orCount := len(g.DependsOn.OrConditions)
		andCount := len(g.DependsOn.AndConditions)

		if orCount < 1 && andCount < 1 {
			return false
		}

		orMet := false
		for _, condition := range g.DependsOn.OrConditions {
			if condition.Match() {
				orMet = true
				break
			}
		}

		andMetCount := 0
		for _, condition := range g.DependsOn.AndConditions {
			if condition.Match() {
				andMetCount++
			}
		}

		if orCount > 0 && andCount < 1 {
			return orMet
		}
		if orCount < 1 && andCount > 0 {
			return andCount == andMetCount
		}
		if orCount > 0 && andCount > 0 {
			return orMet && (andMetCount == orCount)
		}
		return false
	})

	return group
}

type DependsOn struct {
	AndConditions []Condition `yaml:"and_conditions"  json:"and_conditions" mapstructure:"and_conditions"`
	OrConditions  []Condition `yaml:"or_conditions"   json:"or_conditions"  mapstructure:"or_conditions"`
}

type Condition struct {
	fields map[FieldKey]huh.Field

	ParameterName    string      `yaml:"parameter_name"       json:"parameter_name"       mapstructure:"parameter_name"`
	ValueEmpty       *bool       `yaml:"value_empty"          json:"value_empty"          mapstructure:"value_empty"`
	ValueEquals      interface{} `yaml:"value_equals"         json:"value_equals"         mapstructure:"value_equals"`
	ValueNotEquals   interface{} `yaml:"value_not_equals"     json:"value_not_equals"     mapstructure:"value_not_equals"`
	ValueContains    interface{} `yaml:"value_contains"       json:"value_contains"       mapstructure:"value_contains"`
	ValueNotContains interface{} `yaml:"value_not_contains"   json:"value_not_contains"   mapstructure:"value_not_contains"`
}

func (c Condition) Validate() []error {
	var errs []error
	if strutil.IsEmpty(c.ParameterName) {
		errs = append(errs, errors.NewMissingErr("parameter_name", "condition"))
	}
	if c.ValueEmpty == nil && c.ValueEquals == nil && c.ValueNotEquals == nil &&
		c.ValueNotContains == nil && c.ValueContains == nil {
		errs = append(errs, standarderrs.New("missing a valid condition"))
	}
	return errs
}

func (c Condition) Match() bool {
	if c.ValueEmpty != nil {
		return c.IsEmpty(*c.ValueEmpty)
	}
	if c.ValueEquals != nil {
		return c.Equal()
	}
	if c.ValueNotEquals != nil {
		return c.NotEqual()
	}
	if c.ValueContains != nil {
		return c.Contains()
	}
	if c.ValueNotContains != nil {
		return c.NotContains()
	}
	return false
}

func (c Condition) Equal() bool {
	val := c.fields[GetFiledKey(c.ParameterName)]
	return helpers.Equal(c.ValueEquals, val)
}

func (c Condition) NotEqual() bool {
	val := c.fields[GetFiledKey(c.ParameterName)]
	return helpers.NotEqual(c.ValueNotEquals, val)
}

func (c Condition) Contains() bool {
	val := c.fields[GetFiledKey(c.ParameterName)]
	return helpers.Contains(val, c.ValueContains)
}

func (c Condition) NotContains() bool {
	val := c.fields[GetFiledKey(c.ParameterName)]
	return helpers.NotContains(val, c.ValueNotContains)
}

func (c Condition) IsEmpty(empty bool) bool {
	val, ok := c.fields[GetFiledKey(c.ParameterName)]
	if !ok {
		return false
	}
	if empty && helpers.Empty(val) {
		return true
	}
	if !empty && helpers.NotEmpty(val) {
		return true
	}
	return false
}
