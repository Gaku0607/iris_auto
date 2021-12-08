package process

import (
	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/iris_auto/model"
)

type TripartiteForm struct {
	tf model.TripartiteFormParms
}

func NewTripartiteForm() *TripartiteForm {
	return &TripartiteForm{tf: model.Environment.TF}
}

func (t *TripartiteForm) TripartiteForm(c *augo.Context) {
	c.Get(model.SOURCE_KEY)
}
