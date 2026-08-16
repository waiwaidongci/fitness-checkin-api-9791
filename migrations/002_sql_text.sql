CREATE TABLE IF NOT EXISTS workout_notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workout_id INTEGER NOT NULL,
    note TEXT NOT NULL CHECK (note <> ';'),
    FOREIGN KEY(workout_id) REFERENCES workouts(id)
);
