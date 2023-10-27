package collector

import "fmt"

func createTableQuery(projectName, experimentName string) string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s-%s_log (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        
        timestamp INTEGER,
        source TEXT,
        line TEXT,
        
        experiment_name TEXT,
        project_name TEXT,
        run_name TEXT,
        container_id TEXT,
        node_rank TEXT,

        stop_or_err TEXT,

        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`, projectName, experimentName)
}

func createIndexQuery(projectName, experimentName string) string {
	pe := projectName + "-" + experimentName
	return fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s_log_index ON %s_log (
    timestamp, source, 

    project_name, experiment_name, run_name, container_id, node_rank, 

    run_name, stop_or_err)`,
		pe,
		pe,
	)
}

func insertLogQuery(projectName, experimentName string) string {
	return fmt.Sprintf(
		`INSERT INTO %s-%s_log (
    timestamp, 
    source, 
    line, 
    experiment_name, 
    project_name, 
    run_name, 
    container_id, 
    node_rank
  ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, projectName, experimentName)
}

func insertStopQuery(projectName, experimentName string) string {
	return fmt.Sprintf(
		`INSERT INTO %s-%s_log (
    experiment_name, 
    project_name, 
    run_name, 
    container_id, 
    node_rank, 
    stop_or_err
  ) VALUES ($1, $2, $3, $4, $5, $6)`, projectName, experimentName)
}

func checkIfTableExistsQuery(projectName, experimentName string) string {
	return fmt.Sprintf(`SELECT name FROM sqlite_master WHERE type='table' AND name='%s-%s_log'`, projectName, experimentName)
}

func listAllExistingTables() string {
	return `SELECT name FROM sqlite_master WHERE type='table'`
}
