// Code generated by entc, DO NOT EDIT.

package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/dexidp/dex/storage"
	"github.com/dexidp/dex/storage/ent/db/keys"
	"gopkg.in/square/go-jose.v2"
)

// KeysCreate is the builder for creating a Keys entity.
type KeysCreate struct {
	config
	mutation *KeysMutation
	hooks    []Hook
}

// SetVerificationKeys sets the "verification_keys" field.
func (kc *KeysCreate) SetVerificationKeys(sk []storage.VerificationKey) *KeysCreate {
	kc.mutation.SetVerificationKeys(sk)
	return kc
}

// SetSigningKey sets the "signing_key" field.
func (kc *KeysCreate) SetSigningKey(jwk jose.JSONWebKey) *KeysCreate {
	kc.mutation.SetSigningKey(jwk)
	return kc
}

// SetSigningKeyPub sets the "signing_key_pub" field.
func (kc *KeysCreate) SetSigningKeyPub(jwk jose.JSONWebKey) *KeysCreate {
	kc.mutation.SetSigningKeyPub(jwk)
	return kc
}

// SetNextRotation sets the "next_rotation" field.
func (kc *KeysCreate) SetNextRotation(t time.Time) *KeysCreate {
	kc.mutation.SetNextRotation(t)
	return kc
}

// SetID sets the "id" field.
func (kc *KeysCreate) SetID(s string) *KeysCreate {
	kc.mutation.SetID(s)
	return kc
}

// Mutation returns the KeysMutation object of the builder.
func (kc *KeysCreate) Mutation() *KeysMutation {
	return kc.mutation
}

// Save creates the Keys in the database.
func (kc *KeysCreate) Save(ctx context.Context) (*Keys, error) {
	var (
		err  error
		node *Keys
	)
	if len(kc.hooks) == 0 {
		if err = kc.check(); err != nil {
			return nil, err
		}
		node, err = kc.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*KeysMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = kc.check(); err != nil {
				return nil, err
			}
			kc.mutation = mutation
			if node, err = kc.sqlSave(ctx); err != nil {
				return nil, err
			}
			mutation.id = &node.ID
			mutation.done = true
			return node, err
		})
		for i := len(kc.hooks) - 1; i >= 0; i-- {
			if kc.hooks[i] == nil {
				return nil, fmt.Errorf("db: uninitialized hook (forgotten import db/runtime?)")
			}
			mut = kc.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, kc.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX calls Save and panics if Save returns an error.
func (kc *KeysCreate) SaveX(ctx context.Context) *Keys {
	v, err := kc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (kc *KeysCreate) Exec(ctx context.Context) error {
	_, err := kc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (kc *KeysCreate) ExecX(ctx context.Context) {
	if err := kc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (kc *KeysCreate) check() error {
	if _, ok := kc.mutation.VerificationKeys(); !ok {
		return &ValidationError{Name: "verification_keys", err: errors.New(`db: missing required field "verification_keys"`)}
	}
	if _, ok := kc.mutation.SigningKey(); !ok {
		return &ValidationError{Name: "signing_key", err: errors.New(`db: missing required field "signing_key"`)}
	}
	if _, ok := kc.mutation.SigningKeyPub(); !ok {
		return &ValidationError{Name: "signing_key_pub", err: errors.New(`db: missing required field "signing_key_pub"`)}
	}
	if _, ok := kc.mutation.NextRotation(); !ok {
		return &ValidationError{Name: "next_rotation", err: errors.New(`db: missing required field "next_rotation"`)}
	}
	if v, ok := kc.mutation.ID(); ok {
		if err := keys.IDValidator(v); err != nil {
			return &ValidationError{Name: "id", err: fmt.Errorf(`db: validator failed for field "id": %w`, err)}
		}
	}
	return nil
}

func (kc *KeysCreate) sqlSave(ctx context.Context) (*Keys, error) {
	_node, _spec := kc.createSpec()
	if err := sqlgraph.CreateNode(ctx, kc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return nil, err
	}
	return _node, nil
}

func (kc *KeysCreate) createSpec() (*Keys, *sqlgraph.CreateSpec) {
	var (
		_node = &Keys{config: kc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: keys.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeString,
				Column: keys.FieldID,
			},
		}
	)
	if id, ok := kc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := kc.mutation.VerificationKeys(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeJSON,
			Value:  value,
			Column: keys.FieldVerificationKeys,
		})
		_node.VerificationKeys = value
	}
	if value, ok := kc.mutation.SigningKey(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeJSON,
			Value:  value,
			Column: keys.FieldSigningKey,
		})
		_node.SigningKey = value
	}
	if value, ok := kc.mutation.SigningKeyPub(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeJSON,
			Value:  value,
			Column: keys.FieldSigningKeyPub,
		})
		_node.SigningKeyPub = value
	}
	if value, ok := kc.mutation.NextRotation(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: keys.FieldNextRotation,
		})
		_node.NextRotation = value
	}
	return _node, _spec
}

// KeysCreateBulk is the builder for creating many Keys entities in bulk.
type KeysCreateBulk struct {
	config
	builders []*KeysCreate
}

// Save creates the Keys entities in the database.
func (kcb *KeysCreateBulk) Save(ctx context.Context) ([]*Keys, error) {
	specs := make([]*sqlgraph.CreateSpec, len(kcb.builders))
	nodes := make([]*Keys, len(kcb.builders))
	mutators := make([]Mutator, len(kcb.builders))
	for i := range kcb.builders {
		func(i int, root context.Context) {
			builder := kcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*KeysMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				nodes[i], specs[i] = builder.createSpec()
				var err error
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, kcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, kcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{err.Error(), err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, kcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (kcb *KeysCreateBulk) SaveX(ctx context.Context) []*Keys {
	v, err := kcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (kcb *KeysCreateBulk) Exec(ctx context.Context) error {
	_, err := kcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (kcb *KeysCreateBulk) ExecX(ctx context.Context) {
	if err := kcb.Exec(ctx); err != nil {
		panic(err)
	}
}
