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
	"fmt"

	"github.com/google/uuid"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"

	"github.com/self-host/self-host/api/aapije/rest"
	"github.com/self-host/self-host/postgres"
)

// ProgramService represents the repository used for interacting with Program records.
type ProgramService struct {
	q  *postgres.Queries
	db *sql.DB
}

// NewDatasetService instantiates the ProgramService repository.
func NewProgramService(db *sql.DB) *ProgramService {
	if db == nil {
		return nil
	}

	return &ProgramService{
		q:  postgres.New(db),
		db: db,
	}
}

type AddProgramParams struct {
	Name      string
	Type      string
	State     string
	Schedule  string
	Deadline  int
	Language  string
	CreatedBy uuid.UUID
	Tags      []string
}

func (s *ProgramService) AddProgram(ctx context.Context, p AddProgramParams) (*rest.Program, error) {
	params := postgres.CreateProgramParams{
		Name:      p.Name,
		Type:      p.Type,
		State:     p.State,
		Schedule:  p.Schedule,
		Deadline:  int32(p.Deadline),
		Language:  p.Language,
		CreatedBy: p.CreatedBy,
		Tags:      p.Tags,
	}

	program, err := s.q.CreateProgram(ctx, params)
	if err != nil {
		return nil, err
	}

	v := &rest.Program{
		Uuid:     program.Uuid.String(),
		Name:     program.Name,
		Type:     rest.ProgramType(program.Type),
		State:    rest.ProgramState(program.State),
		Schedule: program.Schedule,
		Deadline: int(program.Deadline),
		Language: rest.ProgramLanguage(program.Language),
		Tags:     p.Tags,
	}

	return v, nil
}

type AddCodeRevisionParams struct {
	ProgramUuid uuid.UUID
	CreatedBy   uuid.UUID
	Code        []byte
}

func (s *ProgramService) AddCodeRevision(ctx context.Context, p AddCodeRevisionParams) (*rest.CodeRevision, error) {
	params := postgres.CreateCodeRevisionParams{
		ProgramUuid: p.ProgramUuid,
		CreatedBy:   p.CreatedBy,
		Code:        p.Code,
	}

	rev, err := s.q.CreateCodeRevision(ctx, params)
	if err != nil {
		return nil, err
	}

	v := &rest.CodeRevision{
		Revision:  int(rev.Revision),
		Created:   rev.Created,
		CreatedBy: rev.CreatedBy.String(),
		Checksum:  string(rev.Checksum),
	}

	if rev.Signed.Valid {
		v.Signed = &rev.Signed.Time
	}
	if rev.SignedBy != NilUUID {
		u := rev.SignedBy.String()
		v.SignedBy = &u
	}

	return v, nil
}

func (s *ProgramService) FindAll(ctx context.Context, p FindAllParams) ([]*rest.Program, error) {
	programs := make([]*rest.Program, 0)

	params := postgres.FindProgramsParams{
		Token: p.Token,
	}

	if p.Limit.Value != 0 {
		params.ArgLimit = p.Limit.Value
	}
	if p.Offset.Value != 0 {
		params.ArgOffset = p.Offset.Value
	}

	programs_list, err := s.q.FindPrograms(ctx, params)
	if err != nil {
		return nil, err
	} else {
		for _, t := range programs_list {
			program := &rest.Program{
				Uuid:     t.Uuid.String(),
				Name:     t.Name,
				Type:     rest.ProgramType(t.Type),
				State:    rest.ProgramState(t.State),
				Schedule: t.Schedule,
				Deadline: int(t.Deadline),
				Language: rest.ProgramLanguage(t.Language),
				Tags:     t.Tags,
			}

			programs = append(programs, program)
		}
	}

	return programs, nil
}

func (svc *ProgramService) FindByTags(ctx context.Context, p FindByTagsParams) ([]*rest.Program, error) {
	programs := make([]*rest.Program, 0)

	params := postgres.FindProgramsByTagsParams{
		Tags:  p.Tags,
		Token: p.Token,
	}
	if p.Limit.Value != 0 {
		params.ArgLimit = p.Limit.Value
	}
	if p.Offset.Value != 0 {
		params.ArgOffset = p.Offset.Value
	}

	prog_list, err := svc.q.FindProgramsByTags(ctx, params)
	if err != nil {
		return nil, err
	} else {
		for _, t := range prog_list {
			program := &rest.Program{
				Uuid:     t.Uuid.String(),
				Name:     t.Name,
				Type:     rest.ProgramType(t.Type),
				State:    rest.ProgramState(t.State),
				Schedule: t.Schedule,
				Deadline: int(t.Deadline),
				Language: rest.ProgramLanguage(t.Language),
				Tags:     t.Tags,
			}

			programs = append(programs, program)
		}
	}

	return programs, nil
}

func (s *ProgramService) FindProgramByUuid(ctx context.Context, id uuid.UUID) (*rest.Program, error) {
	program, err := s.q.FindProgramByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	v := &rest.Program{
		Uuid:     program.Uuid.String(),
		Name:     program.Name,
		Type:     rest.ProgramType(program.Type),
		State:    rest.ProgramState(program.State),
		Schedule: program.Schedule,
		Deadline: int(program.Deadline),
		Language: rest.ProgramLanguage(program.Language),
		Tags:     program.Tags,
	}

	return v, nil
}

func (s *ProgramService) FindAllCodeRevisions(ctx context.Context, id uuid.UUID) ([]*rest.CodeRevision, error) {
	revisions := make([]*rest.CodeRevision, 0)

	rev_list, err := s.q.FindProgramCodeRevisions(ctx, id)
	if err != nil {
		return nil, err
	} else {
		for _, t := range rev_list {
			rev := &rest.CodeRevision{
				Revision:  int(t.Revision),
				Created:   t.Created,
				CreatedBy: t.CreatedBy.String(),
				Checksum:  string(t.Checksum),
			}

			if t.Signed.Valid {
				v := t.Signed.Time
				rev.Signed = &v
			}
			if t.SignedBy != NilUUID {
				u := t.SignedBy.String()
				rev.SignedBy = &u
			}

			revisions = append(revisions, rev)
		}
	}

	return revisions, nil
}

func (s *ProgramService) DiffProgramCodeAtRevisions(ctx context.Context, id uuid.UUID, revA int, revB int) (string, error) {
	var codeA, codeB string

	if revA == -1 {
		cA, err := s.q.GetProgramCodeAtHead(ctx, id)
		if err != nil {
			return "", err
		}
		codeA = string(cA.Code)
		revA = int(cA.Revision)
	} else {
		cA, err := s.q.GetProgramCodeAtRevision(ctx, postgres.GetProgramCodeAtRevisionParams{
			ProgramUuid: id,
			Revision:    int32(revA),
		})
		if err != nil {
			return "", err
		}
		codeA = string(cA)
	}

	if revB == -1 {
		cB, err := s.q.GetProgramCodeAtHead(ctx, id)
		if err != nil {
			return "", err
		}
		codeB = string(cB.Code)
		revB = int(cB.Revision)
	} else {
		cB, err := s.q.GetProgramCodeAtRevision(ctx, postgres.GetProgramCodeAtRevisionParams{
			ProgramUuid: id,
			Revision:    int32(revB),
		})
		if err != nil {
			return "", err
		}
		codeB = string(cB)
	}

	aName := fmt.Sprintf("%v@%v", id.String(), revA)
	bName := fmt.Sprintf("%v@%v", id.String(), revB)
	edits := myers.ComputeEdits(span.URIFromPath(id.String()), codeA, codeB)
	return fmt.Sprint(gotextdiff.ToUnified(aName, bName, codeA, edits)), nil
}

func (s *ProgramService) GetProgramCodeAtHead(ctx context.Context, id uuid.UUID) (string, error) {
	row, err := s.q.GetProgramCodeAtHead(ctx, id)
	if err != nil {
		return "", err
	}
	return string(row.Code), nil
}

func (s *ProgramService) GetSignedProgramCodeAtHead(ctx context.Context, id uuid.UUID) (string, error) {
	row, err := s.q.GetSignedProgramCodeAtHead(ctx, id)
	if err != nil {
		return "", err
	}
	return string(row.Code), nil
}

type UpdateProgramByUuidParams struct {
	Name     *string
	Type     *string
	State    *string
	Schedule *string
	Deadline *int
	Language *string
	Tags     *[]string
}

func (s *ProgramService) UpdateProgramByUuid(ctx context.Context, id uuid.UUID, p UpdateProgramByUuidParams) (int64, error) {
	// Use a transaction for this action
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}

	var count int64

	q := s.q.WithTx(tx)

	if p.Name != nil {
		c, err := q.SetProgramNameByUUID(ctx, postgres.SetProgramNameByUUIDParams{
			Uuid: id,
			Name: *p.Name,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Type != nil {
		c, err := q.SetProgramTypeByUUID(ctx, postgres.SetProgramTypeByUUIDParams{
			Uuid: id,
			Type: *p.Type,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.State != nil {
		c, err := q.SetProgramStateByUUID(ctx, postgres.SetProgramStateByUUIDParams{
			Uuid:  id,
			State: *p.State,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Schedule != nil {
		c, err := q.SetProgramScheduleByUUID(ctx, postgres.SetProgramScheduleByUUIDParams{
			Uuid:     id,
			Schedule: *p.Schedule,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Deadline != nil {
		c, err := q.SetProgramDeadlineByUUID(ctx, postgres.SetProgramDeadlineByUUIDParams{
			Uuid:     id,
			Deadline: int32(*p.Deadline),
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Language != nil {
		c, err := q.SetProgramLanguageByUUID(ctx, postgres.SetProgramLanguageByUUIDParams{
			Uuid:     id,
			Language: *p.Language,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		} else {
			count += c
		}
	}

	if p.Tags != nil {
		params := postgres.SetProgramTagsParams{
			Uuid: id,
			Tags: *p.Tags,
		}
		c, err := q.SetProgramTags(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		count += c
	}

	tx.Commit()

	return count, nil
}

type SignCodeRevisionParams struct {
	ProgramUuid uuid.UUID
	Revision    int
	SignedBy    uuid.UUID
}

func (s *ProgramService) SignCodeRevision(ctx context.Context, p SignCodeRevisionParams) (int64, error) {
	count, err := s.q.SignProgramCodeRevision(ctx, postgres.SignProgramCodeRevisionParams{
		ProgramUuid: p.ProgramUuid,
		Revision:    int32(p.Revision),
		SignedBy:    p.SignedBy,
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *ProgramService) DeleteProgram(ctx context.Context, id uuid.UUID) (int64, error) {
	count, err := s.q.DeleteProgram(ctx, id)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *ProgramService) DeleteProgramCodeRevision(ctx context.Context, id uuid.UUID, revision int) (int64, error) {
	count, err := s.q.DeleteProgramCodeRevision(ctx, postgres.DeleteProgramCodeRevisionParams{
		ProgramUuid: id,
		Revision:    int32(revision),
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}
