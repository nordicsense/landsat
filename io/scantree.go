package io

import (
	"io/ioutil"
	"path"
	"regexp"
)

func ScanTree(root, pattern string) ([]string, error) {
	matcher, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return scanWithMatcher(root, matcher)
}

func scanWithMatcher(root string, matcher *regexp.Regexp) ([]string, error) {
	var res []string
	infos, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		fpath := path.Join(root, info.Name())
		if info.IsDir() {
			more, err := scanWithMatcher(fpath, matcher)
			if err != nil {
				return nil, err
			}
			res = append(res, more...)
		} else if matcher.MatchString(info.Name()) {
			res = append(res, fpath)
		}
	}
	return res, nil
}
