package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"wplatform/backend/internal/backup"
	"wplatform/backend/internal/models"
	"wplatform/backend/internal/project"
)

type BackupScheduler struct {
	cron          *cron.Cron
	backupService *backup.BackupService
	scheduleRepo  *backup.BackupScheduleRepository
	projectRepo   *project.ProjectRepository
}

func NewBackupScheduler(
	backupService *backup.BackupService,
	scheduleRepo *backup.BackupScheduleRepository,
	projectRepo *project.ProjectRepository,
) *BackupScheduler {
	return &BackupScheduler{
		cron:          cron.New(cron.WithSeconds()),
		backupService: backupService,
		scheduleRepo:  scheduleRepo,
		projectRepo:   projectRepo,
	}
}

func (s *BackupScheduler) Start() {
	s.cron.AddFunc("@every 5m", func() {
		s.RunScheduledBackups(context.Background())
	})

	s.cron.Start()
	log.Println("Backup scheduler started")
}

func (s *BackupScheduler) Stop() {
	s.cron.Stop()
	log.Println("Backup scheduler stopped")
}

func (s *BackupScheduler) RunScheduledBackups(ctx context.Context) {
	schedules, err := s.scheduleRepo.ListEnabled(ctx)
	if err != nil {
		log.Printf("Failed to list enabled backup schedules: %v", err)
		return
	}

	for _, schedule := range schedules {
		if s.ShouldRunBackup(schedule) {
			req := &backup.BackupRequest{
				ProjectID: schedule.ProjectID,
				Type:      "scheduled",
			}

			_, err := s.backupService.CreateBackup(ctx, req)
			if err != nil {
				log.Printf("Failed to create scheduled backup for project %s: %v", schedule.ProjectID, err)
			} else {
				log.Printf("Created scheduled backup for project %s", schedule.ProjectID)
			}

			if schedule.RetentionDays > 0 {
				if err := s.backupService.ApplyRetentionPolicy(ctx, schedule.ProjectID, schedule.RetentionDays); err != nil {
					log.Printf("Failed to apply retention policy for project %s: %v", schedule.ProjectID, err)
				}
			}
		}
	}
}

func (s *BackupScheduler) ShouldRunBackup(schedule *models.BackupSchedule) bool {
	now := time.Now()

	cronSchedule, err := cron.ParseStandard(schedule.ScheduleCron)
	if err != nil {
		log.Printf("Invalid cron schedule for project %s: %v", schedule.ProjectID, err)
		return false
	}

	lastRun := schedule.UpdatedAt
	nextRun := cronSchedule.Next(lastRun)

	return now.After(nextRun) || now.Equal(nextRun)
}
