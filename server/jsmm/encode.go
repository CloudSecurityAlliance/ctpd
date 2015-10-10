//    Copyright 2015 Cloud Security Alliance EMEA (cloudsecurityalliance.org)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package jsmm

import (
	"reflect"
    "strconv"
    "errors"
)

const (
    TAG = "jsmm"
)

func importValue(val reflect.Value) (MachineValue, error) {
    typ := val.Type()

    switch val.Kind() {
    case reflect.Ptr:
        if val.IsNil() {
            return NewNull(), nil
        }
        return importValue(val.Elem())
    case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
        return NewNumber(float64(val.Uint())), nil
    case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
        return NewNumber(float64(val.Int())), nil
    case reflect.Float32, reflect.Float64:
        return NewNumber(float64(val.Float())), nil
    case reflect.Bool:
        return NewBoolean(val.Bool()), nil
    case reflect.String:
        return NewString(val.String()), nil
    case reflect.Slice, reflect.Array:
        array := NewObject()
        for i := 0; i< val.Len(); i++ {
            v, e := importValue(val.Index(i))
            if e!=nil {
                return NewNull(), e
            }
            array.SetProperty(strconv.Itoa(i),v)
        }
        return array, nil
    case reflect.Struct:
        object := NewObject()
        for i := 0; i< val.NumField(); i++ {
            if val.Field(i).CanInterface() {
                name := ""
                fieldinfo := typ.Field(i)

                if n := fieldinfo.Tag.Get(TAG); n!=""  {
                    if n!="-" {
                        name = n
                    }
                } else {
                    name = fieldinfo.Name
                }

                if name!="" {
                    v, e := importValue(val.Field(i))
                    if e!=nil {
                        return NewNull(), e
                    }
                    object.SetProperty(name,v)
                }
            }
        }
        return object, nil
    case reflect.Map:
        if typ.Key().Kind() == reflect.String {
            object := NewObject()
            for _, key := range val.MapKeys() {
                v, e := importValue(val.MapIndex(key))
                if e!=nil {
                    return NewNull(), e
                }
                object.SetProperty(key.String(),v)
            }
            return object, nil
        }
        return NewNull(), errors.New("Cannot import maps with non-string indices")
    case reflect.Interface:
        return importValue(val.Elem())
    }
    return NewNull(), errors.New("Cannot import " + typ.String())
}

func Import(item interface{}) (MachineValue, error) {
	return importValue(reflect.ValueOf(item))
}

func ImportGlobal(m *Machine, name string, item interface{}) error {
    v, err := Import(item)
    if err!=nil {
        return err
    }
    m.GlobalObject().SetProperty(name,v)
    return nil
}
