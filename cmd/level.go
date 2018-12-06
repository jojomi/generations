package main

type LevelConfig struct {
	Absolute    []AbsoluteLevel `yaml:"absolute,omitempty"`
	Relative    []RelativeLevel `yaml:"relative,omitempty"`
	Combined    []AbsoluteLevel `yaml:"-"`
	BoxOptions  LevelBoxOptions `yaml:"box-options,omitempty"`
	ColorScheme string          `yaml:"color-scheme,omitempty"`
}

func (l *LevelConfig) AddDefaultLevels(from, to int) *LevelConfig {
	existingLevels := map[int]struct{}{}
	for _, level := range l.Absolute {
		existingLevels[level.Index] = struct{}{}
	}
	for i := from; i <= to; i++ {
		if _, ok := existingLevels[i]; !ok {
			l.Absolute = append(l.Absolute, AbsoluteLevel{Index: i})
		}
	}
	return l
}

func (l *LevelConfig) SetColorScheme(colorScheme LevelConfig) *LevelConfig {
	return l
}

func (l *LevelConfig) SetGlobalBoxOptions() *LevelConfig {
	for i, level := range l.Absolute {
		if level.BoxOptions.IsEmpty() {
			level.BoxOptions = l.BoxOptions
		}
		l.Absolute[i] = level
	}
	return l
}

func (l *LevelConfig) Combine(probandLevel int) {
	combined := make([]AbsoluteLevel, 0, len(l.Absolute))
outer:
	for _, a := range l.Absolute {
		baseConfig := a
		for _, r := range l.Relative {
			if baseConfig.Index != r.Index+probandLevel {
				continue
			}
			combined = append(combined, Merge(baseConfig, r))
			continue outer
		}
		combined = append(combined, baseConfig)
	}
	l.Combined = combined
}

func (l *LevelConfig) Inherit(probandLevel int, baseConfig LevelConfig) {
outerAbs:
	for i, a := range l.Absolute {
		for _, b := range baseConfig.Absolute {
			if a.Index != b.Index {
				continue
			}
			if !a.Color.IsEmpty() {
				a.Color = b.Color
			}
			if !a.BoxOptions.IsEmpty() {
				a.BoxOptions = b.BoxOptions
			}
			l.Absolute[i] = a
			continue outerAbs
		}
	}

outerRel:
	for i, r := range l.Relative {
		for _, b := range baseConfig.Relative {
			if b.Index != r.Index+probandLevel {
				continue
			}
			if !r.Color.IsEmpty() {
				r.Color = b.Color
			}
			if !r.BoxOptions.IsEmpty() {
				r.BoxOptions = b.BoxOptions
			}
			l.Relative[i] = r
			continue outerRel
		}
	}
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
