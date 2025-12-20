package main

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/config"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/db"
	zapLogger "github.com/JanArsMAI/Trafic-Incident-Service.git/logger"

	"github.com/brianvoe/gofakeit/v6"
	"go.uber.org/zap"
)

func main() {
	var (
		seed       = flag.Int64("seed", time.Now().UnixNano(), "rng seed")
		usersN     = flag.Int("users", 10000, "users count")
		driversN   = flag.Int("drivers", 10000, "drivers count")
		vehiclesN  = flag.Int("vehicles", 10000, "vehicles count")
		accidentsN = flag.Int("accidents", 10000, "accidents count")
	)
	flag.Parse()

	gofakeit.Seed(*seed)
	rnd := rand.New(rand.NewSource(*seed))

	cfgPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.MustLoad(cfgPath)
	if err != nil {
		panic("config load failed")
	}

	logger := zapLogger.NewLogger(cfg.Logging.Level)
	dbConn, err := db.NewPostgresConnection(db.ReadConfig())
	if err != nil {
		logger.Fatal("db connect failed", zap.Error(err))
	}
	defer dbConn.Close()

	ctx := context.Background()

	logger.Info("seeding started")
	_, _ = dbConn.ExecContext(ctx, `SELECT set_config('app.current_user_id','1',true)`)
	userIDs := make([]int, 0, *usersN)
	for i := 0; i < *usersN; i++ {
		var id int
		err := dbConn.QueryRowContext(ctx, `
			INSERT INTO users (username, password_hash, email, role_id)
			VALUES ($1,$2,$3,$4)
			RETURNING id
		`,
			gofakeit.Username(),
			"hashed_password",
			gofakeit.Email(),
			1+rnd.Intn(3),
		).Scan(&id)
		if err == nil {
			userIDs = append(userIDs, id)
		}
	}
	driverIDs := make([]int, 0, *driversN)
	for i := 0; i < *driversN; i++ {
		exp := rnd.Intn(40)
		var id int
		err := dbConn.QueryRowContext(ctx, `
			INSERT INTO drivers (
				full_name,
				date_of_birth,
				license_number,
				license_issue_date,
				experience_years
			)
			VALUES ($1,$2,$3,$4,$5)
			RETURNING id
		`,
			gofakeit.Name(),
			gofakeit.DateRange(
				time.Now().AddDate(-70, 0, 0),
				time.Now().AddDate(-18, 0, 0),
			),
			gofakeit.Regex(`[A-Z]{2}[0-9]{6}`),
			time.Now().AddDate(-exp, 0, 0),
			exp,
		).Scan(&id)
		if err == nil {
			driverIDs = append(driverIDs, id)
		}
	}
	vehicleIDs := make([]int, 0, *vehiclesN)
	for i := 0; i < *vehiclesN; i++ {
		var id int
		err := dbConn.QueryRowContext(ctx, `
			INSERT INTO vehicles (
				plate_number,
				model,
				year,
				vehicle_type,
				owner_driver_id
			)
			VALUES ($1,$2,$3,$4,$5)
			RETURNING id
		`,
			gofakeit.Regex(`[A-Z][0-9]{3}[A-Z]{2}`),
			gofakeit.CarModel(),
			rnd.Intn(25)+1998,
			[]string{"sedan", "truck", "bus", "motorcycle"}[rnd.Intn(4)],
			driverIDs[rnd.Intn(len(driverIDs))],
		).Scan(&id)
		if err == nil {
			vehicleIDs = append(vehicleIDs, id)
		}
	}

	weatherIDs := make([]int, 0, 200)
	for i := 0; i < 200; i++ {
		var id int
		_ = dbConn.QueryRowContext(ctx, `
			INSERT INTO weather (
				temperature,
				precipitation,
				visibility,
				road_condition,
				description
			)
			VALUES ($1,$2,$3,$4,$5)
			RETURNING id
		`,
			gofakeit.Float64Range(-20, 35),
			[]string{"none", "rain", "snow", "fog"}[rnd.Intn(4)],
			rnd.Intn(1000),
			[]string{"dry", "wet", "icy"}[rnd.Intn(3)],
			gofakeit.Sentence(6),
		).Scan(&id)
		weatherIDs = append(weatherIDs, id)
	}
	inspectorIDs := make([]int, 0, 100)
	for i := 0; i < 100; i++ {
		var id int
		_ = dbConn.QueryRowContext(ctx, `
			INSERT INTO inspectors (
				full_name,
				badge_number,
				department,
				rank,
				user_id
			)
			VALUES ($1,$2,$3,$4,$5)
			RETURNING id
		`,
			gofakeit.Name(),
			gofakeit.UUID(),
			"ГИБДД",
			[]string{"лейтенант", "капитан", "майор"}[rnd.Intn(3)],
			userIDs[rnd.Intn(len(userIDs))],
		).Scan(&id)
		inspectorIDs = append(inspectorIDs, id)
	}
	for i := 0; i < *accidentsN; i++ {
		var accidentID int
		_ = dbConn.QueryRowContext(ctx, `
			INSERT INTO accidents (
				location,
				date_time,
				weather_id,
				inspector_id,
				severity
			)
			VALUES ($1,$2,$3,$4,$5)
			RETURNING id
		`,
			gofakeit.City(),
			gofakeit.DateRange(time.Now().AddDate(-5, 0, 0), time.Now()),
			weatherIDs[rnd.Intn(len(weatherIDs))],
			inspectorIDs[rnd.Intn(len(inspectorIDs))],
			[]string{"low", "medium", "high", "fatal"}[rnd.Intn(4)],
		).Scan(&accidentID)

		participants := 1 + rnd.Intn(3)
		for p := 0; p < participants; p++ {
			driverID := driverIDs[rnd.Intn(len(driverIDs))]
			vehicleID := vehicleIDs[rnd.Intn(len(vehicleIDs))]

			var partID int
			err := dbConn.QueryRowContext(ctx, `
				INSERT INTO accident_participants (
					accident_id,
					driver_id,
					vehicle_id,
					is_guilty,
					injuries
				)
				VALUES ($1,$2,$3,$4,$5)
				ON CONFLICT DO NOTHING
				RETURNING id
			`,
				accidentID,
				driverID,
				vehicleID,
				p == 0,
				[]string{"none", "light", "serious"}[rnd.Intn(3)],
			).Scan(&partID)

			if err == nil {
				_, _ = dbConn.ExecContext(ctx, `
					INSERT INTO penalties (
						participant_id,
						amount,
						issued_by,
						status
					)
					VALUES ($1,$2,$3,$4)
				`,
					partID,
					rnd.Intn(50000)+1000,
					inspectorIDs[rnd.Intn(len(inspectorIDs))],
					[]string{"paid", "unpaid", "canceled"}[rnd.Intn(3)],
				)
			}
		}
	}

	logger.Info("seeding finished successfully")
}
