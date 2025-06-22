package env

// import (
// 	"errors"
// 	"fmt"
// 	"strings"
// 	"unicode/utf8"

// 	"github.com/ardnew/envmux/pkg"
// )

// var (
// 	errKeyEmpty     = errors.New("empty key")
// 	errKeyInvalid   = errors.New("invalid key")
// 	errKeyReserved  = errors.New("reserved key")
// 	errKeyDuplicate = errors.New("duplicate key")
// 	errKeyUndefined = errors.New("undefined key")
// )

// type mapError struct {
// 	error
// 	key string
// 	val any
// }

// func (e mapError) Error() string {
// 	return fmt.Sprintf("map error: %q: %v", e.key, e.error)
// }

// func (e mapError) Is(target error) bool {
// 	if _, ok := target.(mapError); ok {
// 		return true
// 	}
// 	return errors.Is(e.error, target)
// }

// type filterKeyFunc[T any] func(key string, exclude ...venv[T]) (string,
// error)

// func uniqueKey[T any](key string, exclude ...venv[T]) (string, error) {
// 	key = strings.TrimSpace(key)
// 	switch {
// 	case key == "":
// 		return key, pkg.Errorf("%w: %q", errKeyEmpty, key)
// 	case !utf8.ValidString(key):
// 		return key, pkg.Errorf("%w: %q", errKeyInvalid, key)
// 	default:
// 		for _, e := range exclude {
// 			if e == nil {
// 				continue
// 			}
// 			if val, ok := e[key]; ok {
// 				return key, duplicateKeyError{key: key, val: val}
// 			}
// 		}
// 	}
// 	return key, nil
// }

// func replaceKey[T any](key string, _ ...venv[T]) (string, error) {
// 	key, err := uniqueKey[T](key)
// 	if errors.Is(err, duplicateKeyError{}) {
// 		err = nil // ignore duplicate key error
// 	}
// 	return key, err
// }

// func (f filterKeyFunc[T]) reduce(envs ...venv[T]) pkg.Option[venv[T]] {
// 	return func(env venv[T]) venv[T] {
// 		if env == nil {
// 			env = make(venv[T])
// 		}
// 		for _, e := range envs {
// 			if e == nil {
// 				continue
// 			}
// 			for k, v := range e {
// 				key, err := f(k)
// 				if err != nil {
// 					continue
// 				}
// 				env[key] = v
// 			}
// 		}
// 		return env
// 	}
// }

// func (f filterKeyFunc[T]) with(include venv[T], exclude ...venv[T])
// pkg.Option[venv[T]] {
// 	return func(env venv[T]) venv[T] {
// 		if env == nil {
// 			env = make(venv[T])
// 		}
// 		if include == nil {
// 			return env
// 		}
// 		for k, v := range include {
// 			key, err := f(k, exclude...)
// 			if err != nil {
// 				continue
// 			}
// 			env[key] = v
// 		}
// 		return env
// 	}
// }

// func (v venv[T]) del(key ...string) (venv[T], error) {
// 	for _, k := range key {
// 		k, err := uniqueKey[T](k)
// 		switch {
// 		case err == nil:
// 			continue // undefined key, nothing to do
// 		case errors.Is(err, errKeyEmpty):
// 			continue // empty key, nothing to do
// 		case errors.Is(err, errKeyInvalid):
// 			return v, err // error if trying to delete invalid key
// 		case errors.Is(err, errKeyReserved):
// 			return v, err // error if trying to delete reserved key
// 		case errors.Is(err, duplicateKeyError{}):
// 			delete(v, k) // found key, proceed to delete
// 		}
// 	}
// 	return v, nil
// }

// func (v venv[T]) env() venv[string] {
// 	env := make(venv[string])
// 	for key, val := range v {
// 		switch v := any(val).(type) {
// 		case string:
// 			env[key] = v
// 		default:
// 			if s, ok := v.(fmt.Stringer); ok {
// 				env[key] = s.String()
// 			} else if e, ok := v.(error); ok {
// 				env[key] = e.Error()
// 			} else {
// 				env[key] = fmt.Sprintf("%v", v)
// 			}
// 		}
// 	}
// 	return env
// }
