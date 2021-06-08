/*
Copyright (C) 2021 The Self-host Authors.
This file is part of Self-host <https://github.com/self-host/self-host>.

Self-host is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Self-host is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Self-host.  If not, see <http://www.gnu.org/licenses/>.
*/

package services

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/self-host/self-host/api/selfserv/rest"
	pg "github.com/self-host/self-host/postgres"
)

// ThingService represents the repository used for interacting with Thing records.
type ThingService struct {
	q  *pg.Queries
	db *sql.DB
}

// NewThingService instantiates the ThingService repository.
func NewThingService(db *sql.DB) *ThingService {
	if db == nil {
		return nil
	}

	return &ThingService{
		q:  pg.New(db),
		db: db,
	}
}

func (svc *ThingService) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	found, err := svc.q.ExistsThing(ctx, id)
	if err != nil {
		return false, err
	}

	return found > 0, nil
}

func (svc *ThingService) AddThing(ctx context.Context, name string, thing_type *string, created_by *uuid.UUID) (*rest.Thing, error) {
	// Use a transaction for this action
	tx, err := svc.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		// Log?
		return nil, err
	}

	q := svc.q.WithTx(tx)

	params := pg.CreateThingParams{
		Name: name,
	}

	if thing_type != nil {
		params.Type.Scan(*thing_type)
	}

	if created_by != nil {
		params.CreatedBy = *created_by
	}

	thing, err := q.CreateThing(ctx, params)
	if err != nil {
		tx.Rollback()
		return nil, err
	} else {
		tx.Commit()
	}

	v := &rest.Thing{
		Uuid:      thing.Uuid.String(),
		Name:      thing.Name,
		CreatedBy: thing.CreatedBy.String(),
		State:     rest.ThingState(thing.State),
	}

	if thing.Type.Valid {
		v.Type = &thing.Type.String
	}

	return v, nil
}

func (svc *ThingService) FindThingByUuid(ctx context.Context, thing_uuid uuid.UUID) (*rest.Thing, error) {
	t, err := svc.q.FindThingByUUID(ctx, thing_uuid)
	if err != nil {
		return nil, err
	}

	thing := &rest.Thing{
		Uuid:      t.Uuid.String(),
		Name:      t.Name,
		State:     rest.ThingState(t.State),
		CreatedBy: t.CreatedBy.String(),
	}

	if t.Type.Valid {
		thing.Type = &t.Type.String
	}

	return thing, nil
}

func (svc *ThingService) FindAll(ctx context.Context, token []byte, limit *int64, offset *int64) ([]*rest.Thing, error) {
	things := make([]*rest.Thing, 0)

	params := pg.FindThingsParams{
		Token:     token,
		ArgLimit:  20,
		ArgOffset: 0,
	}
	if limit != nil {
		params.ArgLimit = *limit
	}
	if offset != nil {
		params.ArgOffset = *offset
	}

	thing_list, err := svc.q.FindThings(ctx, params)
	if err != nil {
		return nil, err
	} else {
		for _, t := range thing_list {
			thing := &rest.Thing{
				Uuid:      t.Uuid.String(),
				Name:      t.Name,
				State:     rest.ThingState(t.State),
				CreatedBy: t.CreatedBy.String(),
			}
			if t.Type.Valid {
				thing.Type = &t.Type.String
			}

			things = append(things, thing)
		}
	}

	return things, nil
}

func (svc *ThingService) UpdateByUuid(ctx context.Context, id uuid.UUID, name *string, thingtype *string, state *string) (int64, error) {
	// Use a transaction for this action
	tx, err := svc.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}

	q := svc.q.WithTx(tx)

	var count int64

	if name != nil {
		params := pg.SetThingNameByUUIDParams{
			Uuid: id,
			Name: *name,
		}
		c, err := q.SetThingNameByUUID(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if thingtype != nil {
		var ns sql.NullString
		ns.Scan(thingtype)

		params := pg.SetThingTypeByUUIDParams{
			Uuid: id,
			Type: ns,
		}
		c, err := q.SetThingTypeByUUID(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	if state != nil {
		params := pg.SetThingStateByUUIDParams{
			Uuid:  id,
			State: pg.ThingState(*state),
		}
		c, err := q.SetThingStateByUUID(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	tx.Commit()

	return count, nil
}

func (svc *ThingService) DeleteThing(ctx context.Context, thing_uuid uuid.UUID) (int64, error) {
	count, err := svc.q.DeleteThing(ctx, thing_uuid)
	if err != nil {
		return 0, err
	}

	return count, nil
}
