package main

import (
	"fmt"

	"github.com/jojomi/go-spew/spew"
)

type Level interface {
	GetColor() LevelColor
	GetBoxOptions() LevelBoxOptions
	SetColor(c LevelColor)
	SetBoxOptions(b LevelBoxOptions)
}

type LevelConfig struct {
	Absolute []AbsoluteLevel `yaml:"absolute,omitempty"`
	Relative []RelativeLevel `yaml:"relative,omitempty"`

	// globals
	BoxOptions LevelBoxOptions `yaml:"box-options,omitempty"`
	Color      LevelColor      `yaml:"color,omitempty"`

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

func (l *LevelConfig) SetColorScheme(colorScheme LevelConfig) *LevelConfig {
	return l
}

func (l *LevelConfig) SetGlobals() *LevelConfig {
	for i, level := range l.Absolute {
		level.BoxOptions = *level.BoxOptions.Merge(l.BoxOptions)
		level.Color = *level.Color.Merge(l.Color)
		fmt.Println(level.Index)
		fmt.Println(l.Color, level.Color)
		spew.Dump(level)
		l.Absolute[i] = level
	}
	///spew.Dump(l)
	///os.Exit(6)
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
			Merge(baseConfig, r)
			combined = append(combined, baseConfig)
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
			Merge(a, b)
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
			Merge(r, b)
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

func (lc *LevelColor) Equals(other LevelColor) bool {
	return lc.Main == other.Main && lc.Leaf == other.Leaf
}

func (lc *LevelColor) IsEmpty() bool {
	return lc.Main == "" && lc.Leaf == ""
}

func (lc *LevelColor) Merge(inheritedColor LevelColor) *LevelColor {
	spew.Dump(lc)
	spew.Dump(inheritedColor)
	if lc.Main == "" {
		lc.Main = inheritedColor.Main
	}
	if lc.Leaf == "" {
		lc.Leaf = inheritedColor.Leaf
	}
	if lc.Leaf != "" || inheritedColor.Leaf != "" {
		///os.Exit(6)
	}
	return lc
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

func (lo *LevelBoxOptions) Merge(inheritedOpts LevelBoxOptions) *LevelBoxOptions {
	if lo.Main != "" {
		lo.Main = inheritedOpts.Main
	}
	if lo.Leaf != "" {
		lo.Leaf = inheritedOpts.Leaf
	}
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

func (l AbsoluteLevel) GetColor() LevelColor {
	return l.Color
}

func (l AbsoluteLevel) SetColor(lc LevelColor) {
	l.Color = lc
}

func (l AbsoluteLevel) GetBoxOptions() LevelBoxOptions {
	return l.BoxOptions
}

func (l AbsoluteLevel) SetBoxOptions(lb LevelBoxOptions) {
	l.BoxOptions = lb
}

func (l RelativeLevel) GetColor() LevelColor {
	return l.Color
}

func (l RelativeLevel) SetColor(lc LevelColor) {
	l.Color = lc
}

func (l RelativeLevel) GetBoxOptions() LevelBoxOptions {
	return l.BoxOptions
}

func (l RelativeLevel) SetBoxOptions(lb LevelBoxOptions) {
	l.BoxOptions = lb
}

func Merge(a, b Level) {
	aColor := a.GetColor()
	aColor.Merge(b.GetColor())
	a.SetColor(aColor)

	aBoxOptions := a.GetBoxOptions()
	aBoxOptions.Merge(b.GetBoxOptions())
	a.SetBoxOptions(aBoxOptions)
}
