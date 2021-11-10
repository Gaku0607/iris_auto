package slicer

import (
	"strings"

	"github.com/Gaku0607/excelgo"
)

type MomoCut struct {
	sf Slicer //大小比較切割
	cf Slicer //內容值切割
}

func NewMomoCut(sf Slicer, cf Slicer) *MomoCut {
	m := &MomoCut{}
	m.sf = sf
	m.cf = cf
	return m
}

func (c *MomoCut) Cut(s *excelgo.Sourc, rows [][]string) ([][]string, [][]string, error) {
	//target為（ＭＯＭＯ指定）
	other, target, err := c.cf.Cut(s, rows)
	if err != nil {
		return nil, nil, err
	}

	bigs, normals, err := c.sf.Cut(s, other)
	if err != nil {
		return nil, nil, err
	}

	col := s.GetCol("注文番号").Col
	var ts [][]string

	for _, t := range bigs {

		num := t[col]
		index := strings.Index(num, "-")
		for i := 0; i < len(normals); i++ {
			if strings.Contains(normals[i][col], num[:index]) {
				ts = append(ts, normals[i])
				normals = append(normals[:i], normals[i+1:]...)
				i--
			}
		}

	}
	bigs = append(bigs, ts...)

	target = append(target, normals...)

	return bigs, target, nil
}

func (m *MomoCut) Status() int {
	return SLICER_MOMO
}
