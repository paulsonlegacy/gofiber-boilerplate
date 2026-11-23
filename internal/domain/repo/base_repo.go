// base_repo.go — generic repository that every concrete repo embeds.
// Provides standard CRUD and read methods so concrete repos only define domain-specific queries.
package repo

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)


// ============================================================
// HELPERS
// ============================================================

// Preload describes a single GORM association to load with a query.
type Preload struct {
	Entity string
	Limit  int    // 0 = no limit
	Order  string // "" = no ordering
}

func applyPreload(db *gorm.DB, p Preload) *gorm.DB {
	if p.Limit == 0 && p.Order == "" {
		return db.Preload(p.Entity)
	}
	return db.Preload(p.Entity, func(db *gorm.DB) *gorm.DB {
		if p.Order != "" {
			db = db.Order(p.Order)
		}
		if p.Limit > 0 {
			db = db.Limit(p.Limit)
		}
		return db
	})
}

func applyORConditions(db *gorm.DB, conditions map[string]any) *gorm.DB {
	if len(conditions) == 0 {
		return db
	}
	first := true
	for col, val := range conditions {
		clause := fmt.Sprintf("%s = ?", col)
		if first {
			db = db.Where(clause, val)
			first = false
		} else {
			db = db.Or(clause, val)
		}
	}
	return db
}

// applyORConditionsNormalized builds a GORM OR query normalizing string values to lowercase.
// Use only when all condition values are strings (e.g. email, code lookups).
func applyORConditionsNormalized(q *gorm.DB, conditions map[string]any) *gorm.DB {
	if len(conditions) == 0 {
		return q
	}
	var result *gorm.DB
	for col, val := range conditions {
		strVal, ok := val.(string)
		if !ok {
			clause := fmt.Sprintf("%s = ?", col)
			if result == nil {
				result = q.Where(clause, val)
			} else {
				result = result.Or(clause, val)
			}
			continue
		}
		clause := fmt.Sprintf("LOWER(TRIM(%s)) = ?", col)
		if result == nil {
			result = q.Where(clause, strings.ToLower(strings.TrimSpace(strVal)))
		} else {
			result = result.Or(clause, strings.ToLower(strings.TrimSpace(strVal)))
		}
	}
	return result
}

func normalizePagination(page, limit int, order string) (int, int, string) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if order == "" {
		order = "created_at DESC"
	}
	return page, limit, order
}

// applySelect chains .Select(fields) when fields is non-empty, otherwise returns db unchanged.
// Used internally by every read method — callers pass nil/empty to get SELECT *.
func applySelect(db *gorm.DB, fields []string) *gorm.DB {
	if len(fields) == 0 {
		return db
	}
	return db.Select(fields)
}


// ============================================================
// BASE REPOSITORY
// ============================================================

// BaseRepository provides generic CRUD and read methods for any GORM entity.
// Embed it in every concrete repository to inherit the standard method set.
//
// All read methods accept an optional trailing `fields ...string` parameter.
// When supplied, only those columns are SELECTed. Omit it to get all columns.
//
//	// All columns (existing behavior — nothing changes)
//	user, err := repo.Get(ctx, map[string]any{"id": id})
//
//	// Only the columns you need
//	user, err := repo.Get(ctx, map[string]any{"email": email}, "id", "email", "password", "status")
//
// Usage:
//
//	type UserRepository struct {
//	    BaseRepository[entity.User]
//	}
type BaseRepository[T any] struct {
	DB *gorm.DB
}


// ============================================================
// WRITES
// ============================================================

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Create(entity).Error
}

func (r *BaseRepository[T]) CreateTx(tx *gorm.DB, entity *T) error {
	return tx.Create(entity).Error
}

func (r *BaseRepository[T]) UpdateFields(ctx context.Context, id any, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return r.DB.WithContext(ctx).
		Model(new(T)).
		Where("id = ?", id).
		Updates(fields).Error
}

func (r *BaseRepository[T]) UpdateFieldsTx(tx *gorm.DB, id any, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return tx.Model(new(T)).
		Where("id = ?", id).
		Updates(fields).Error
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id any) error {
	result := r.DB.WithContext(ctx).Delete(new(T), "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *BaseRepository[T]) DeleteTx(tx *gorm.DB, id any) error {
	result := tx.Delete(new(T), "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}


// ============================================================
// READ — AND conditions
// ============================================================

// Get returns the first T matching all conditions (AND).
// Pass field names to project specific columns; omit for SELECT *.
func (r *BaseRepository[T]) Get(ctx context.Context, conditions map[string]any, fields ...string) (*T, error) {
	var result T
	err := applySelect(r.DB.WithContext(ctx), fields).Where(conditions).First(&result).Error
	return &result, err
}

// GetAll returns all T matching all conditions (AND).
// Pass field names to project specific columns; omit for SELECT *.
func (r *BaseRepository[T]) GetAll(ctx context.Context, conditions map[string]any, fields ...string) ([]*T, error) {
	var items []*T
	err := applySelect(r.DB.WithContext(ctx), fields).Where(conditions).Find(&items).Error
	return items, err
}

// GetAllPaged returns paginated T matching all conditions (AND).
// Count query always uses SELECT * internally — the select only affects the data fetch.
// Pass field names to project specific columns; omit for SELECT *.
func (r *BaseRepository[T]) GetAllPaged(ctx context.Context, conditions map[string]any, page, limit int, order string, fields ...string) ([]*T, int64, error) {
	page, limit, order = normalizePagination(page, limit, order)
	var total int64
	if err := r.DB.WithContext(ctx).Model(new(T)).Where(conditions).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []*T
	err := applySelect(r.DB.WithContext(ctx), fields).
		Where(conditions).
		Order(order).
		Offset((page-1)*limit).
		Limit(limit).
		Find(&items).Error
	return items, total, err
}

// GetWithPreloads returns the first T matching conditions with the given preloads.
// Pass field names to project specific columns on the root entity; omit for SELECT *.
// Note: select does not apply to preloaded associations — scope those in the Preload definition.
func (r *BaseRepository[T]) GetWithPreloads(ctx context.Context, conditions map[string]any, preloads []Preload, fields ...string) (*T, error) {
	var result T
	db := applySelect(r.DB.WithContext(ctx), fields)
	for _, p := range preloads {
		db = applyPreload(db, p)
	}
	err := db.Where(conditions).First(&result).Error
	return &result, err
}

// GetAllWithPreloads returns all T matching conditions with the given preloads.
// Pass field names to project specific columns on the root entity; omit for SELECT *.
func (r *BaseRepository[T]) GetAllWithPreloads(ctx context.Context, conditions map[string]any, preloads []Preload, fields ...string) ([]*T, error) {
	var items []*T
	db := applySelect(r.DB.WithContext(ctx), fields)
	for _, p := range preloads {
		db = applyPreload(db, p)
	}
	err := db.Where(conditions).Find(&items).Error
	return items, err
}

// GetAllPagedWithPreloads returns paginated T with preloads.
// Pass field names to project specific columns on the root entity; omit for SELECT *.
func (r *BaseRepository[T]) GetAllPagedWithPreloads(ctx context.Context, conditions map[string]any, page, limit int, order string, preloads []Preload, fields ...string) ([]*T, int64, error) {
	page, limit, order = normalizePagination(page, limit, order)
	var total int64
	if err := r.DB.WithContext(ctx).Model(new(T)).Where(conditions).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	db := applySelect(r.DB.WithContext(ctx), fields)
	for _, p := range preloads {
		db = applyPreload(db, p)
	}
	var items []*T
	err := db.Where(conditions).Order(order).Offset((page-1)*limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *BaseRepository[T]) Exists(ctx context.Context, conditions map[string]any) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(new(T)).Where(conditions).Count(&count).Error
	return count > 0, err
}


// ============================================================
// READ — OR conditions
// ============================================================

// GetAny returns the first T matching any condition (OR).
// Pass field names to project specific columns; omit for SELECT *.
func (r *BaseRepository[T]) GetAny(ctx context.Context, conditions map[string]any, fields ...string) (*T, error) {
	var result T
	err := applySelect(applyORConditions(r.DB.WithContext(ctx), conditions), fields).First(&result).Error
	return &result, err
}

// GetAllAny returns all T matching any condition (OR).
// Pass field names to project specific columns; omit for SELECT *.
func (r *BaseRepository[T]) GetAllAny(ctx context.Context, conditions map[string]any, fields ...string) ([]*T, error) {
	var items []*T
	err := applySelect(applyORConditions(r.DB.WithContext(ctx), conditions), fields).Find(&items).Error
	return items, err
}

// GetAllAnyPaged returns paginated T matching any condition (OR).
// Pass field names to project specific columns; omit for SELECT *.
func (r *BaseRepository[T]) GetAllAnyPaged(ctx context.Context, conditions map[string]any, page, limit int, order string, fields ...string) ([]*T, int64, error) {
	page, limit, order = normalizePagination(page, limit, order)
	q := applyORConditions(r.DB.WithContext(ctx).Model(new(T)), conditions)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []*T
	err := applySelect(q, fields).Order(order).Offset((page-1)*limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *BaseRepository[T]) ExistsAny(ctx context.Context, conditions map[string]any) (bool, error) {
	var count int64
	err := applyORConditions(r.DB.WithContext(ctx).Model(new(T)), conditions).Count(&count).Error
	return count > 0, err
}
