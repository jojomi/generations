package generations

type Level interface {
	GetColor() LevelColor
	GetBoxOptions() LevelBoxOptions
	GetOptions() LevelOptions
	SetColor(c LevelColor)
	SetBoxOptions(b LevelBoxOptions)
	SetOptions(b LevelOptions)
}

type LevelConfig struct {
	Absolute []AbsoluteLevel `yaml:"absolute,omitempty"`
	Relative []RelativeLevel `yaml:"relative,omitempty"`

	// globals
	Color      LevelColor      `yaml:"color,omitempty"`
	BoxOptions LevelBoxOptions `yaml:"box-options,omitempty"`
	Options    LevelOptions    `yaml:"options,omitempty"`

	Themes []string `yaml:"themes,omitempty"`

	// generated
	Combined []AbsoluteLevel `yaml:"-"`
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

func (l *LevelConfig) SetGlobals() *LevelConfig {
	for i, level := range l.Absolute {
		level.Color = *level.Color.OverwriteWith(&l.Color) // TODO inherit
		level.BoxOptions = *level.BoxOptions.Merge(l.BoxOptions)
		level.Options = *level.Options.Merge(l.Options)
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
			merged := Merge(&baseConfig, &r).(*AbsoluteLevel)
			combined = append(combined, *merged)
			continue outer
		}
		combined = append(combined, baseConfig)
	}
	l.Combined = combined
}

func (l *LevelConfig) Inherit(probandLevel int, baseConfig LevelConfig) *LevelConfig {
outerAbs:
	for i, a := range l.Absolute {
		for _, b := range baseConfig.Combined {
			if a.Index != b.Index {
				continue
			}
			a = *(MergeWithBase(&a, &b).(*AbsoluteLevel))
			l.Absolute[i] = a
			continue outerAbs
		}
	}

	/*outerRel:
	for i, r := range l.Relative {
		for _, b := range baseConfig.Relative {
			if b.Index != r.Index+probandLevel {
				continue
			}
			r = *(MergeWithBase(&r, &b).(*RelativeLevel))
			l.Relative[i] = r
			continue outerRel
		}
	}*/

	return l
}

type AbsoluteLevel struct {
	Index      int             `yaml:"index,omitempty"`
	Color      LevelColor      `yaml:"color,omitempty"`
	BoxOptions LevelBoxOptions `yaml:"box-options,omitempty"`
	Options    LevelOptions    `yaml:"options,omitempty"`
}

type RelativeLevel struct {
	Index      int
	Color      LevelColor
	BoxOptions LevelBoxOptions `yaml:"box-options,omitempty"`
	Options    LevelOptions    `yaml:"options,omitempty"`
}

type LevelColor struct {
	Main string `yaml:"main,omitempty"`
	Leaf string `yaml:"leaf,omitempty"`
}

func (lc *LevelColor) Equals(other LevelColor) bool {
	return lc.Main == other.Main && lc.Leaf == other.Leaf
}

func (lc *LevelColor) IsEmpty() bool {
	return lc.Main == "" && lc.Leaf == ""
}

// overwrite
func (lc *LevelColor) OverwriteWith(lc2 *LevelColor) *LevelColor {
	result := &LevelColor{
		Main: lc2.Main,
		Leaf: lc2.Leaf,
	}
	return result
}

// merge with default
func (lc *LevelColor) MergeWithBase(lc2 *LevelColor) *LevelColor {
	result := &LevelColor{
		Main: lc2.Main,
		Leaf: lc2.Leaf,
	}
	if lc.Main != "" {
		result.Main = lc.Main
	}
	if lc.Leaf != "" {
		result.Leaf = lc.Leaf
	}
	return result
}

type LevelBoxOptions struct {
	Main string `yaml:"main,omitempty"`
	Leaf string `yaml:"leaf,omitempty"`
}

func (lo *LevelBoxOptions) Equals(other LevelBoxOptions) bool {
	return lo.Main == other.Main && lo.Leaf == other.Leaf
}

func (lo *LevelBoxOptions) IsEmpty() bool {
	return lo.Main == "" && lo.Leaf == ""
}

// Merge does append!
func (lo *LevelBoxOptions) Merge(inheritedOpts LevelBoxOptions) *LevelBoxOptions {
	if inheritedOpts.Main != "" && lo.Main != "" {
		inheritedOpts.Main += "%\n%\n"
	}
	lo.Main = inheritedOpts.Main + lo.Main
	if inheritedOpts.Leaf != "" && lo.Leaf != "" {
		inheritedOpts.Leaf += "%\n%\n"
	}
	lo.Leaf = inheritedOpts.Leaf + lo.Leaf
	return lo
}

type LevelOptions string

func (lo *LevelOptions) Merge(inheritedOpts LevelOptions) *LevelOptions {
	if inheritedOpts != "" && *lo != "" {
		inheritedOpts += "%\n%\n"
	}
	result := LevelOptions(string(inheritedOpts) + string(*lo))
	lo = &result
	return lo
}

func (l *AbsoluteLevel) Equals(other AbsoluteLevel) bool {
	return l.Color.Equals(other.Color) &&
		l.BoxOptions.Equals(other.BoxOptions)
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

func (l *AbsoluteLevel) GetColor() LevelColor {
	return l.Color
}

func (l *AbsoluteLevel) SetColor(lc LevelColor) {
	l.Color = lc
}

func (l *AbsoluteLevel) GetBoxOptions() LevelBoxOptions {
	return l.BoxOptions
}

func (l *AbsoluteLevel) SetBoxOptions(lb LevelBoxOptions) {
	l.BoxOptions = lb
}

func (l *AbsoluteLevel) GetOptions() LevelOptions {
	return l.Options
}

func (l *AbsoluteLevel) SetOptions(lb LevelOptions) {
	l.Options = lb
}

func (l *RelativeLevel) GetColor() LevelColor {
	return l.Color
}

func (l *RelativeLevel) SetColor(lc LevelColor) {
	l.Color = lc
}

func (l *RelativeLevel) GetBoxOptions() LevelBoxOptions {
	return l.BoxOptions
}

func (l *RelativeLevel) SetBoxOptions(lb LevelBoxOptions) {
	l.BoxOptions = lb
}

func (l *RelativeLevel) GetOptions() LevelOptions {
	return l.Options
}

func (l *RelativeLevel) SetOptions(lb LevelOptions) {
	l.Options = lb
}

func Merge(a, b Level) Level {
	aColor := a.GetColor()
	bColor := b.GetColor()
	resultingColor := aColor.OverwriteWith(&bColor)
	a.SetColor(*resultingColor)

	aBoxOptions := a.GetBoxOptions()
	aBoxOptions.Merge(b.GetBoxOptions())
	a.SetBoxOptions(aBoxOptions)

	aOptions := a.GetOptions()
	bOptions := b.GetOptions()
	resultingOptions := aOptions.Merge(bOptions)
	a.SetOptions(*resultingOptions)
	return a
}

func MergeWithBase(a, b Level) Level {
	aColor := a.GetColor()
	bColor := b.GetColor()
	resultingColor := aColor.MergeWithBase(&bColor)
	a.SetColor(*resultingColor)

	aBoxOptions := a.GetBoxOptions()
	aBoxOptions.Merge(b.GetBoxOptions())
	a.SetBoxOptions(aBoxOptions)

	aOptions := a.GetOptions()
	bOptions := b.GetOptions()
	resultingOptions := aOptions.Merge(bOptions)
	a.SetOptions(*resultingOptions)
	return a
}
