/*
Copyright 2018 The pdfcpu Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pdfcpu

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/log"
	"github.com/pkg/errors"
)

// Dict represents a PDF dict object.
type Dict map[string]Object

// NewDict returns a new PDFDict object.
func NewDict() Dict {
	return map[string]Object{}
}

// Len returns the length of this PDFDict.
func (d Dict) Len() int {
	return len(d)
}

// Insert adds a new entry to this PDFDict.
func (d Dict) Insert(key string, value Object) (ok bool) {
	_, found := d.Find(key)
	if !found {
		d[key] = value
	}
	return true
}

// InsertInt adds a new int entry to this PDFDict.
func (d Dict) InsertInt(key string, value int) {
	d.Insert(key, Integer(value))
}

// InsertFloat adds a new float entry to this PDFDict.
func (d Dict) InsertFloat(key string, value float32) {
	d.Insert(key, Float(value))
}

// InsertString adds a new string entry to this PDFDict.
func (d Dict) InsertString(key, value string) {
	d.Insert(key, StringLiteral(value))
}

// InsertName adds a new name entry to this PDFDict.
func (d Dict) InsertName(key, value string) {
	d.Insert(key, Name(value))
}

// Update modifies an existing entry of this PDFDict.
func (d Dict) Update(key string, value Object) {
	if value != nil {
		d[key] = value
	}
}

// Find returns the Object for given key and PDFDict.
func (d Dict) Find(key string) (value Object, found bool) {
	value, found = d[key]
	return
}

// Delete deletes the Object for given key.
func (d Dict) Delete(key string) (value Object) {

	value, found := d.Find(key)
	if !found {
		return nil
	}

	delete(d, key)

	return
}

// Entry returns the value for given key.
func (d Dict) Entry(dictName, key string, required bool) (Object, error) {
	obj, found := d.Find(key)
	if !found || obj == nil {
		if required {
			return nil, errors.Errorf("dict=%s required entry=%s missing", dictName, key)
		}
		//log.Trace.Printf("dict=%s entry %s is nil\n", dictName, key)
		return nil, nil
	}
	return obj, nil
}

// BooleanEntry expects and returns a BooleanEntry for given key.
func (d Dict) BooleanEntry(key string) (value bool, ok bool) {

	v, found := d.Find(key)
	if !found {
		return false, false
	}

	bb, ok := v.(Boolean)
	if ok {
		return bb.Value(), true
	}

	return false, false
}

// StringEntry expects and returns a StringLiteral entry for given key.
// Unused.
func (d Dict) StringEntry(key string) (string, bool) {

	value, found := d.Find(key)
	if !found {
		return "", false
	}

	pdfStr, ok := value.(StringLiteral)
	if ok {
		return string(pdfStr), true
	}

	return "", false
}

// NameEntry expects and returns a Name entry for given key.
func (d Dict) NameEntry(key string) (string, bool) {

	value, found := d.Find(key)
	if !found {
		return "", false
	}

	Name, ok := value.(Name)
	if ok {
		return string(Name), true
	}

	return "", false
}

// IntEntry expects and returns a Integer entry for given key.
func (d Dict) IntEntry(key string) (int, bool) {

	value, found := d.Find(key)
	if !found {
		return 0, false
	}

	pdfInt, ok := value.(Integer)
	if ok {
		return int(pdfInt), true
	}

	return 0, false
}

// Int64Entry expects and returns a Integer entry representing an int64 value for given key.
func (d Dict) Int64Entry(key string) (int64, bool) {

	value, found := d.Find(key)
	if !found {
		return 0, false
	}

	pdfInt, ok := value.(Integer)
	if ok {
		return int64(pdfInt), true
	}

	return 0, false
}

// IndirectRefEntry returns an indirectRefEntry for given key for this dictionary.
func (d Dict) IndirectRefEntry(key string) (ref IndirectRef, ok bool) {

	value, found := d.Find(key)
	if !found {
		return ref, false
	}

	ref, ok = value.(IndirectRef)
	return ref, ok
}

// DictEntry expects and returns a PDFDict entry for given key.
func (d Dict) DictEntry(key string) (Dict, bool) {

	value, found := d.Find(key)
	if !found {
		return nil, false
	}

	// TODO resolve indirect ref.

	d, ok := value.(Dict)
	if ok {
		return d, true
	}

	return nil, false
}

// StreamDictEntry expects and returns a StreamDict entry for given key.
// unused.
func (d Dict) StreamDictEntry(key string) (dict StreamDict, ok bool) {

	value, found := d.Find(key)
	if !found {
		return dict, false
	}

	dict, ok = value.(StreamDict)
	return dict, ok
}

// ArrayEntry expects and returns a Array entry for given key.
func (d Dict) ArrayEntry(key string) (array Array, ok bool) {

	value, found := d.Find(key)
	if !found {
		return array, false
	}

	array, ok = value.(Array)
	return array, ok
}

// StringLiteralEntry returns a StringLiteral object for given key.
func (d Dict) StringLiteralEntry(key string) (s StringLiteral, ok bool) {

	value, found := d.Find(key)
	if !found {
		return s, false
	}

	s, ok = value.(StringLiteral)
	return s, ok
}

// HexLiteralEntry returns a HexLiteral object for given key.
func (d Dict) HexLiteralEntry(key string) (hex HexLiteral, ok bool) {

	value, found := d.Find(key)
	if !found {
		return hex, false
	}

	hex, ok = value.(HexLiteral)
	return hex, ok
}

// Length returns a int64 for entry with key "Length".
// Stream length may be referring to an indirect object.
func (d Dict) Length() (int64, int, bool) {

	val, ok := d.Int64Entry("Length")
	if ok {
		return val, 0, true
	}

	indirectRef, ok := d.IndirectRefEntry("Length")
	if !ok {
		return 0, 0, false
	}

	intVal := indirectRef.ObjectNumber.Value()

	return 0, intVal, true
}

// Type returns the value of the name entry for key "Type".
func (d Dict) Type() string {
	s, _ := d.NameEntry("Type")
	return s
}

// Subtype returns the value of the name entry for key "Subtype".
func (d Dict) Subtype() string {
	s, _ := d.NameEntry("Subtype")
	return s
}

// Size returns the value of the int entry for key "Size"
func (d Dict) Size() int {
	i, _ := d.IntEntry("Size")
	return i
}

// IsObjStm returns true if given PDFDict is an object stream.
func (d Dict) IsObjStm() bool {
	return d.Type() == "ObjStm"
}

// W returns a Array for key "W".
func (d Dict) W() (Array, bool) {
	return d.ArrayEntry("W")
}

// Prev returns the previous offset.
func (d Dict) Prev() int64 {
	i, _ := d.Int64Entry("Prev")
	return i
}

// Index returns a *Array for key "Index".
func (d Dict) Index() (Array, bool) {
	return d.ArrayEntry("Index")
}

// N returns a *int for key "N".
func (d Dict) N() int {
	i, _ := d.IntEntry("N")
	return i
}

// First returns a int for key "First".
func (d Dict) First() (int, bool) {
	return d.IntEntry("First")
}

// IsLinearizationParmDict returns true if this dict has an int entry for key "Linearized".
func (d Dict) IsLinearizationParmDict() bool {
	_, ok := d.IntEntry("Linearized")
	return ok
}

// IncrementBy increments the integer value for given key by i.
func (d *Dict) IncrementBy(key string, i int) error {
	v, ok := d.IntEntry(key)
	if !ok {
		return errors.Errorf("IncrementBy: unknown key: %s", key)
	}
	v += i
	d.Update(key, Integer(v))
	return nil
}

// Increment increments the integer value for given key.
func (d *Dict) Increment(key string) error {
	return d.IncrementBy(key, 1)
}

func (d Dict) indentedString(level int) string {

	logstr := []string{"<<\n"}
	tabstr := strings.Repeat("\t", level)

	var keys []string
	for k := range d {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {

		v := d[k]

		if subdict, ok := v.(Dict); ok {
			dictStr := subdict.indentedString(level + 1)
			logstr = append(logstr, fmt.Sprintf("%s<%s, %s>\n", tabstr, k, dictStr))
			continue
		}

		if a, ok := v.(Array); ok {
			arrStr := a.indentedString(level + 1)
			logstr = append(logstr, fmt.Sprintf("%s<%s, %s>\n", tabstr, k, arrStr))
			continue
		}

		logstr = append(logstr, fmt.Sprintf("%s<%s, %v>\n", tabstr, k, v))

	}

	logstr = append(logstr, fmt.Sprintf("%s%s", strings.Repeat("\t", level-1), ">>"))

	return strings.Join(logstr, "")
}

// PDFString returns a string representation as found in and written to a PDF file.
func (d Dict) PDFString() string {

	logstr := []string{} //make([]string, 20)
	logstr = append(logstr, "<<")

	var keys []string
	for k := range d {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {

		v := d[k]

		if v == nil {
			logstr = append(logstr, fmt.Sprintf("/%s null", k))
			continue
		}

		d, ok := v.(Dict)
		if ok {
			logstr = append(logstr, fmt.Sprintf("/%s%s", k, d.PDFString()))
			continue
		}

		a, ok := v.(Array)
		if ok {
			logstr = append(logstr, fmt.Sprintf("/%s%s", k, a.PDFString()))
			continue
		}

		ir, ok := v.(IndirectRef)
		if ok {
			logstr = append(logstr, fmt.Sprintf("/%s %s", k, ir.PDFString()))
			continue
		}

		n, ok := v.(Name)
		if ok {
			logstr = append(logstr, fmt.Sprintf("/%s%s", k, n.PDFString()))
			continue
		}

		i, ok := v.(Integer)
		if ok {
			logstr = append(logstr, fmt.Sprintf("/%s %s", k, i.PDFString()))
			continue
		}

		f, ok := v.(Float)
		if ok {
			logstr = append(logstr, fmt.Sprintf("/%s %s", k, f.PDFString()))
			continue
		}

		b, ok := v.(Boolean)
		if ok {
			logstr = append(logstr, fmt.Sprintf("/%s %s", k, b.PDFString()))
			continue
		}

		sl, ok := v.(StringLiteral)
		if ok {
			logstr = append(logstr, fmt.Sprintf("/%s%s", k, sl.PDFString()))
			continue
		}

		hl, ok := v.(HexLiteral)
		if ok {
			logstr = append(logstr, fmt.Sprintf("/%s%s", k, hl.PDFString()))
			continue
		}

		log.Info.Fatalf("PDFDict.PDFString(): entry of unknown object type: %T %[1]v\n", v)
	}

	logstr = append(logstr, ">>")
	return strings.Join(logstr, "")
}

func (d Dict) String() string {
	return d.indentedString(1)
}

// StringEntryBytes returns the byte slice representing the string value for key.
func (d Dict) StringEntryBytes(key string) ([]byte, error) {

	s, ok := d.StringLiteralEntry(key)
	if ok {
		return Unescape(s.Value())
	}

	h, ok := d.HexLiteralEntry(key)
	if ok {
		return h.Bytes()
	}

	return nil, nil
}
