// Code generated from Pkl module `jmann345.snek.Opts`. DO NOT EDIT.
package opts

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Opts struct {
	Rows int `pkl:"rows"`

	Cols int `pkl:"cols"`

	Fg int `pkl:"fg"`

	Bg int `pkl:"bg"`

	SnekFg int `pkl:"snekFg"`

	FoodFg int `pkl:"foodFg"`

	SnekSkin string `pkl:"snekSkin"`

	FoodSkin string `pkl:"foodSkin"`

	Speed int `pkl:"speed"`

	Snax int `pkl:"snax"`

	Portals bool `pkl:"portals"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Opts
func LoadFromPath(ctx context.Context, path string) (ret *Opts, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := evaluator.Close()
		if err == nil {
			err = cerr
		}
	}()
	ret, err = Load(ctx, evaluator, pkl.FileSource(path))
	return ret, err
}

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Opts
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*Opts, error) {
	var ret Opts
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
