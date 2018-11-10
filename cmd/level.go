package main

type LevelConfig struct {
	Absolute []AbsoluteLevel `yaml:"absolute,omitempty"`
	Relative []RelativeLevel `yaml:"relative,omitempty"`
	Combined []AbsoluteLevel `yaml:"-"`
}

func (l *LevelConfig) Combine(probandLevel int, baseConfig LevelConfig) []AbsoluteLevel {
	combined := make([]AbsoluteLevel, 0, len(l.Absolute))
outer:
	for _, b := range baseConfig.Absolute {
		baseConfig := b
		for _, a := range l.Absolute {
			if a.Index == b.Index {
				baseConfig = a
			}
		}
		for _, r := range l.Relative {
			if baseConfig.Index != r.Index+probandLevel {
				continue
			}
			combined = append(combined, Merge(baseConfig, r))
			continue outer
		}
		combined = append(combined, baseConfig)
	}
	return combined
}

type AbsoluteLevel struct {
	Index      int             `yaml:"index,omitempty"`
	Color      LevelColor      `yaml:"color,omitempty"`
	BoxOptions LevelBoxOptions `yaml:"box-options,omitempty"`
	Options    string          `yaml:"options,omitempty"`
}

type RelativeLevel struct {
	Index      int
	Color      LevelColor
	BoxOptions LevelBoxOptions `yaml:"box-options,omitempty"`
	Options    string          `yaml:"options,omitempty"`
}

type LevelColor struct {
	Main string `yaml:"main,omitempty"`
	Leaf string `yaml:"leaf,omitempty"`
}

func (lc *LevelColor) IsEmpty() bool {
	return lc.Main == "" && lc.Leaf == ""
}

type LevelBoxOptions struct {
	Main string `yaml:"main,omitempty"`
	Leaf string `yaml:"leaf,omitempty"`
}

func (lo *LevelBoxOptions) IsEmpty() bool {
	return lo.Main == "" && lo.Leaf == ""
}

func (l *AbsoluteLevel) IsParentLevel(probandLevel int) bool {
	return l.Index > probandLevel
}

func (l *AbsoluteLevel) IsChildLevel(probandLevel int) bool {
	return l.Index < probandLevel
}

func (l *AbsoluteLevel) IsProbandLevel(probandLevel int) bool {
	return l.Index == probandLevel
}

func Merge(a AbsoluteLevel, r RelativeLevel) AbsoluteLevel {
	if !r.Color.IsEmpty() {
		a.Color = r.Color
	}
	if !r.BoxOptions.IsEmpty() {
		a.BoxOptions = r.BoxOptions
	}
	if r.Options != "" {
		a.Options = r.Options
	}
	return a
}
