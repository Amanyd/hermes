package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/eulerbutcooler/hermes/services/hermes-core/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

type RelayStore struct {
	db *pgxpool.Pool
}

var ErrRelayNotFound = errors.New("relay not found")
var ErrExecutionNotFound = errors.New("execution not found")

func NewRelayStore(db *pgxpool.Pool) *RelayStore {
	return &RelayStore{db: db}
}

func (s *RelayStore) CreateRelay(ctx context.Context, req models.CreateRelayRequest) (*models.RelayWithActions, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	relayID := uuid.New().String()
	webhookPath := fmt.Sprintf("/hooks/%s", relayID)
	now := time.Now().UTC()

	triggerType := req.TriggerType
	if triggerType == "" {
		triggerType = models.TriggerWebhook
	}
	triggerConfig := req.TriggerConfig
	if triggerConfig == nil {
		triggerConfig = map[string]any{}
	}
	triggerConfigJSON, err := json.Marshal(triggerConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal trigger_config: %w", err)
	}

	var nextRunAt *time.Time
	if triggerType == models.TriggerCron {
		schedule, _ := triggerConfig["schedule"].(string)
		next, err := computeNextRun(schedule, now)
		if err != nil {
			return nil, fmt.Errorf("invalid cron schedule: %w", err)
		}
		nextRunAt = &next
	}

	queryRelay := `
			INSERT INTO relays (id, user_id, name, description, webhook_path, is_active,
			                    trigger_type, trigger_config, next_run_at, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			RETURNING id, user_id, name, description, webhook_path, is_active,
			          trigger_type, trigger_config, created_at, updated_at`

	var relay models.Relay
	var triggerConfigBytes []byte
	var triggerTypeStr string

	err = tx.QueryRow(ctx, queryRelay,
		relayID, req.UserID, req.Name, req.Description, webhookPath, true,
		string(triggerType), triggerConfigJSON, nextRunAt, now, now,
	).Scan(
		&relay.ID, &relay.UserID, &relay.Name, &relay.Description,
		&relay.WebhookPath, &relay.IsActive,
		&triggerTypeStr, &triggerConfigBytes,
		&relay.CreatedAt, &relay.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("insert relay: %w", err)
	}
	relay.TriggerType = models.TriggerType(triggerTypeStr)
	if err := json.Unmarshal(triggerConfigBytes, &relay.TriggerConfig); err != nil {
		return nil, fmt.Errorf("unmarshal trigger_config: %w", err)
	}
	actions := make([]models.RelayAction, 0, len(req.Actions))

	queryAction := `INSERT INTO relay_actions(id,relay_id,action_type, config, order_index,created_at,updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7)
	RETURNING id,relay_id,action_type,config,order_index,created_at,updated_at`

	for _, actionReq := range req.Actions {
		actionID := uuid.New().String()
		configJSON, err := json.Marshal(actionReq.Config)
		if err != nil {
			return nil, fmt.Errorf("marshal action config: %w", err)
		}
		var action models.RelayAction
		var configBytes []byte
		err = tx.QueryRow(ctx, queryAction, actionID, relayID, actionReq.ActionType, configJSON, actionReq.OrderIndex, now, now).Scan(
			&action.ID, &action.RelayID, &action.ActionType, &configBytes, &action.OrderIndex, &action.CreatedAt, &action.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert action: %w", err)
		}
		if err := json.Unmarshal(configBytes, &action.Config); err != nil {
			return nil, fmt.Errorf("unmarshal action config: %w", err)
		}
		actions = append(actions, action)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &models.RelayWithActions{
		Relay:   relay,
		Actions: actions,
	}, nil
}

func (s *RelayStore) GetAllRelays(ctx context.Context, userID string) ([]models.Relay, error) {
	query := `SELECT id, user_id, name, description, webhook_path, is_active,
	                 trigger_type, trigger_config, created_at, updated_at
	          FROM relays
	          WHERE user_id = $1::uuid
	          ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query relays: %w", err)
	}
	defer rows.Close()
	relays := make([]models.Relay, 0)
	for rows.Next() {
		var triggerTypeStr string
		var triggerConfigBytes []byte
		var relay models.Relay
		err := rows.Scan(
			&relay.ID,
			&relay.UserID,
			&relay.Name,
			&relay.Description,
			&relay.WebhookPath,
			&relay.IsActive,
			&triggerTypeStr,
			&triggerConfigBytes,
			&relay.CreatedAt,
			&relay.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan relay: %w", err)
		}
		relay.TriggerType = models.TriggerType(triggerTypeStr)
		if len(triggerConfigBytes) > 0 {
			if err := json.Unmarshal(triggerConfigBytes, &relay.TriggerConfig); err != nil {
				return nil, fmt.Errorf("unmarshal trigger_config: %w", err)
			}
		}
		relays = append(relays, relay)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return relays, nil
}

func (s *RelayStore) GetRelay(ctx context.Context, relayID, userID string) (*models.RelayWithActions, error) {
	queryRelay := `
		SELECT id, user_id, name, description, webhook_path, is_active,
		       trigger_type, trigger_config, created_at, updated_at
		FROM relays
		WHERE id = $1 AND user_id = $2
	`
	var relay models.Relay
	var triggerTypeStr string
	var triggerConfigBytes []byte
	err := s.db.QueryRow(ctx, queryRelay, relayID, userID).Scan(
		&relay.ID,
		&relay.UserID,
		&relay.Name,
		&relay.Description,
		&relay.WebhookPath,
		&relay.IsActive,
		&triggerTypeStr,
		&triggerConfigBytes,
		&relay.CreatedAt,
		&relay.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrRelayNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query relay: %w", err)
	}
	relay.TriggerType = models.TriggerType(triggerTypeStr)
	if len(triggerConfigBytes) > 0 {
		if err := json.Unmarshal(triggerConfigBytes, &relay.TriggerConfig); err != nil {
			return nil, fmt.Errorf("unmarshal trigger_config: %w", err)
		}
	}

	queryActions := `
		SELECT id, relay_id, action_type, config, order_index, created_at, updated_at
		FROM relay_actions
		WHERE relay_id = $1
		ORDER BY order_index ASC
	`

	rows, err := s.db.Query(ctx, queryActions, relayID)
	if err != nil {
		return nil, fmt.Errorf("query actions: %w", err)
	}
	defer rows.Close()

	actions := make([]models.RelayAction, 0)
	for rows.Next() {
		var action models.RelayAction
		var configBytes []byte
		err := rows.Scan(
			&action.ID,
			&action.RelayID,
			&action.ActionType,
			&configBytes,
			&action.OrderIndex,
			&action.CreatedAt,
			&action.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan action: %w", err)
		}

		if err := json.Unmarshal(configBytes, &action.Config); err != nil {
			return nil, fmt.Errorf("unmarshal config: %w", err)
		}

		actions = append(actions, action)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &models.RelayWithActions{
		Relay:   relay,
		Actions: actions,
	}, nil
}

func (s *RelayStore) UpdateRelay(ctx context.Context, relayID, userID string, req models.UpdateRelayRequest) (*models.Relay, error) {
	now := time.Now()
	query := `UPDATE relays SET updated_at = $1`
	args := []any{now}
	argIdx := 2

	if req.Name != nil {
		query += fmt.Sprintf(", name=$%d", argIdx)
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Description != nil {
		query += fmt.Sprintf(", description=$%d", argIdx)
		args = append(args, *req.Description)
		argIdx++
	}
	if req.IsActive != nil {
		query += fmt.Sprintf(", is_active=$%d", argIdx)
		args = append(args, *req.IsActive)
		argIdx++
	}
	if req.TriggerType != nil {
		query += fmt.Sprintf(", trigger_type=$%d", argIdx)
		args = append(args, string(*req.TriggerType))
		argIdx++

		triggerConfig := req.TriggerConfig
		if triggerConfig == nil {
			triggerConfig = map[string]any{}
		}
		configJSON, err := json.Marshal(triggerConfig)
		if err != nil {
			return nil, fmt.Errorf("marshal trigger_config: %w", err)
		}
		query += fmt.Sprintf(", trigger_config=$%d", argIdx)
		args = append(args, configJSON)
		argIdx++

		if *req.TriggerType == models.TriggerCron {
			schedule, _ := triggerConfig["schedule"].(string)
			next, err := computeNextRun(schedule, now)
			if err != nil {
				return nil, fmt.Errorf("invalid cron schedule: %w", err)
			}
			query += fmt.Sprintf(", next_run_at=$%d", argIdx)
			args = append(args, next)
			argIdx++
		} else {
			query += fmt.Sprintf(", next_run_at=$%d", argIdx)
			args = append(args, nil)
			argIdx++
		}
	}

	query += fmt.Sprintf(
		" WHERE id=$%d AND user_id=$%d RETURNING id, user_id, name, description, webhook_path, is_active, trigger_type, trigger_config, created_at, updated_at",
		argIdx, argIdx+1,
	)
	args = append(args, relayID, userID)

	var relay models.Relay
	var triggerTypeStr string
	var triggerConfigBytes []byte
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&relay.ID,
		&relay.UserID,
		&relay.Name,
		&relay.Description,
		&relay.WebhookPath,
		&relay.IsActive,
		&triggerTypeStr,
		&triggerConfigBytes,
		&relay.CreatedAt,
		&relay.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrRelayNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update relay: %w", err)
	}
	relay.TriggerType = models.TriggerType(triggerTypeStr)
	if len(triggerConfigBytes) > 0 {
		if err := json.Unmarshal(triggerConfigBytes, &relay.TriggerConfig); err != nil {
			return nil, fmt.Errorf("unmarshal trigger_config: %w", err)
		}
	}
	return &relay, nil
}

func (s *RelayStore) UpdateRelayActions(ctx context.Context, relayID, userID string, actionInputs []models.CreateRelayActionInput) (*models.RelayWithActions, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var relay models.Relay
	err = tx.QueryRow(ctx,
		`SELECT id, user_id, name, description, webhook_path, is_active, created_at, updated_at, trigger_type, trigger_config
		 FROM relays WHERE id = $1 AND user_id = $2`, relayID, userID,
	).Scan(&relay.ID, &relay.UserID, &relay.Name, &relay.Description,
		&relay.WebhookPath, &relay.IsActive, &relay.CreatedAt, &relay.UpdatedAt, &relay.TriggerType, &relay.TriggerConfig)
	if err == pgx.ErrNoRows {
		return nil, ErrRelayNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query relay: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM relay_actions WHERE relay_id = $1`, relayID)
	if err != nil {
		return nil, fmt.Errorf("delete old actions: %w", err)
	}

	now := time.Now()
	queryAction := `INSERT INTO relay_actions(id, relay_id, action_type, config, order_index, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, relay_id, action_type, config, order_index, created_at, updated_at`

	actions := make([]models.RelayAction, 0, len(actionInputs))
	for _, actionReq := range actionInputs {
		actionID := uuid.New().String()
		configJSON, err := json.Marshal(actionReq.Config)
		if err != nil {
			return nil, fmt.Errorf("marshal action config: %w", err)
		}
		var action models.RelayAction
		var configBytes []byte
		err = tx.QueryRow(ctx, queryAction, actionID, relayID, actionReq.ActionType, configJSON, actionReq.OrderIndex, now, now).Scan(
			&action.ID, &action.RelayID, &action.ActionType, &configBytes, &action.OrderIndex, &action.CreatedAt, &action.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert action: %w", err)
		}
		if err := json.Unmarshal(configBytes, &action.Config); err != nil {
			return nil, fmt.Errorf("unmarshal action config: %w", err)
		}
		actions = append(actions, action)
	}

	_, err = tx.Exec(ctx, `UPDATE relays SET updated_at = $1 WHERE id = $2`, now, relayID)
	if err != nil {
		return nil, fmt.Errorf("update relay timestamp: %w", err)
	}
	relay.UpdatedAt = now

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &models.RelayWithActions{
		Relay:   relay,
		Actions: actions,
	}, nil
}

func (s *RelayStore) DeleteRelay(ctx context.Context, relayID, userID string) error {
	query := `DELETE FROM relays WHERE id = $1 AND user_id= $2`
	result, err := s.db.Exec(ctx, query, relayID, userID)
	if err != nil {
		return fmt.Errorf("delete relay: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrRelayNotFound
	}

	return nil
}

func (s *RelayStore) GetExecutions(ctx context.Context, relayID, userID string, limit int) ([]models.Execution, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT e.id, e.relay_id, e.event_id, e.status, e.trigger_payload, e.error_message, e.started_at, e.finished_at
		FROM executions e
		JOIN relays r ON r.id = e.relay_id
		WHERE e.relay_id = $1 AND r.user_id = $2
		ORDER BY e.started_at DESC
		LIMIT $3
	`

	rows, err := s.db.Query(ctx, query, relayID, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("query executions: %w", err)
	}
	defer rows.Close()

	executions := make([]models.Execution, 0)
	for rows.Next() {
		var ex models.Execution
		var payloadBytes []byte
		var errorMsg *string
		var eventID *string

		err := rows.Scan(
			&ex.ID,
			&ex.RelayID,
			&eventID,
			&ex.Status,
			&payloadBytes,
			&errorMsg,
			&ex.StartedAt,
			&ex.FinishedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan execution: %w", err)
		}

		if eventID != nil {
			ex.EventID = *eventID
		}
		if len(payloadBytes) > 0 {
			if err := json.Unmarshal(payloadBytes, &ex.TriggerPayload); err != nil {
				return nil, fmt.Errorf("unmarshal execution payload: %w", err)
			}
		}
		if errorMsg != nil {
			ex.ErrorMessage = *errorMsg
		}

		executions = append(executions, ex)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("executions rows error: %w", err)
	}

	return executions, nil
}

func (s *RelayStore) GetExecutionSteps(ctx context.Context, executionID, userID string) ([]models.ExecutionStep, error) {
	query := `
		SELECT es.id, es.execution_id, es.order_index, es.action_type, es.status, es.input, es.output, es.error_message, es.started_at, es.finished_at
		FROM execution_steps es
		JOIN executions e ON e.id = es.execution_id
		JOIN relays r ON r.id = e.relay_id
		WHERE es.execution_id = $1 AND r.user_id = $2
		ORDER BY es.order_index ASC
	`

	rows, err := s.db.Query(ctx, query, executionID, userID)
	if err != nil {
		return nil, fmt.Errorf("query execution steps: %w", err)
	}
	defer rows.Close()

	steps := make([]models.ExecutionStep, 0)
	for rows.Next() {
		var step models.ExecutionStep
		var inputBytes []byte
		var outputBytes []byte
		var errorMsg *string

		err := rows.Scan(
			&step.ID,
			&step.ExecutionID,
			&step.OrderIndex,
			&step.ActionType,
			&step.Status,
			&inputBytes,
			&outputBytes,
			&errorMsg,
			&step.StartedAt,
			&step.FinishedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan execution step: %w", err)
		}

		if len(inputBytes) > 0 {
			if err := json.Unmarshal(inputBytes, &step.Input); err != nil {
				return nil, fmt.Errorf("unmarshal execution step input: %w", err)
			}
		}
		if len(outputBytes) > 0 {
			if err := json.Unmarshal(outputBytes, &step.Output); err != nil {
				return nil, fmt.Errorf("unmarshal execution step output: %w", err)
			}
		}
		if errorMsg != nil {
			step.ErrorMessage = *errorMsg
		}

		steps = append(steps, step)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("execution steps rows error: %w", err)
	}

	return steps, nil
}

func (s *RelayStore) DeleteExecution(ctx context.Context, executionID, userID string) error {
	query := `
		DELETE FROM executions
		USING relays
		WHERE executions.id = $1
		  AND executions.relay_id = relays.id
		  AND relays.user_id = $2
	`
	result, err := s.db.Exec(ctx, query, executionID, userID)
	if err != nil {
		return fmt.Errorf("delete execution: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrExecutionNotFound
	}
	return nil
}

func computeNextRun(schedule string, from time.Time) (time.Time, error) {
	if schedule == "" {
		return time.Time{}, fmt.Errorf("cron schedule is empty")
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(schedule)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse cron %q: %w", schedule, err)
	}
	return sched.Next(from), nil
}
