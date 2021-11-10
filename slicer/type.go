package slicer

import (
	"strconv"
	"strings"

	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/iris_auto/model"
)

type Slicer interface {
	Cut(*excelgo.Sourc, [][]string) ([][]string, [][]string, error)
	Status() int
}

const (
	SLICER_SIZE = iota
	SLICER_CONTENT
	SLICER_MOMO
)

type SizeCut struct {
	norm int
}

func NewSizeCut(num int) Slicer {
	n := &SizeCut{}
	n.norm = num
	return n
}

func (n *SizeCut) Cut(s *excelgo.Sourc, rows [][]string) ([][]string, [][]string, error) {

	var (
		big [][]string
	)
	col := s.GetCol(model.Environment.SE.SizeSpan)

	n.sort(rows, col.Col)

	for i, row := range rows {
		size, _ := strconv.Atoi(row[col.Col])
		if n.norm > size {
			big = rows[:i]
			rows = rows[i:]
			break
		}
	}
	return big, rows, nil
}

func (m *SizeCut) sort(rows [][]string, col int) {
	for i := 1; i < len(rows); i++ {
		ii := rows[i]
		index := i - 1
		insertVal, _ := strconv.Atoi(rows[i][col])
		targetVal, _ := strconv.Atoi(rows[index][col])

		for targetVal < insertVal {
			rows[index+1] = rows[index]
			index--
			if index < 0 {
				break
			}
			targetVal, _ = strconv.Atoi(rows[index][col])
		}
		if index != i-1 {
			rows[index+1] = ii
		}
	}
}
func (m *SizeCut) Status() int {
	return SLICER_SIZE
}

type ContentCut struct {
	keywords []string
}

func NewContentCut(kws []string) Slicer {
	c := &ContentCut{}
	c.keywords = kws
	return c
}

func (c *ContentCut) Cut(s *excelgo.Sourc, rows [][]string) ([][]string, [][]string, error) {

	var (
		target [][]string
		col    *excelgo.Col
	)

	col = s.GetCol(model.Environment.SE.ContainsSpan)

	for _, kw := range c.keywords {
		for i := 0; i < len(rows); i++ {
			if b := strings.Contains(rows[i][col.Col], kw); b {
				target = append(target, rows[i])
				rows = append(rows[:i], rows[i+1:]...)
				i--
			}
		}
	}
	return rows, target, nil

}

func (m *ContentCut) Status() int {
	return SLICER_CONTENT
}

type CutFile struct {
	s *excelgo.Sourc
}

func NewCutFile(s *excelgo.Sourc) *CutFile {
	f := &CutFile{}
	f.s = s
	return f
}

func (n *CutFile) Cut(rows [][]string, cf Slicer) ([][]string, [][]string, error) {

	bigs, normals, err := cf.Cut(n.s, rows)
	if err != nil {
		return nil, nil, err
	}
	bigs, normals = n.Merge(bigs, normals, 0)
	return bigs, normals, nil
}

func (n *CutFile) Merge(target, other [][]string, status int) ([][]string, [][]string) {
	col := n.s.GetCol("注文番号").Col
	var (
		ts     [][]string
		nother [][]string
	)
	switch status {
	case SLICER_CONTENT, SLICER_SIZE:
		for _, t := range target {
			val := t[col]
			for i := 0; i < len(other); i++ {
				//宅配（小貨品）句尾可能含有Ａ必須先排除比較
				if val == strings.TrimRight(other[i][col], "A") {
					ts = append(ts, other[i])
					other = append(other[:i], other[i+1:]...)
					i--
				}
			}
		}
	default:
	}

	for _, v := range other {
		nother = append(nother, v)
	}
	return append(target, ts...), nother
}
