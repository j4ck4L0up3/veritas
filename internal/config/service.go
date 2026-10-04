package config

import (
	"charm.land/log/v2"
	"fmt"
	"os"
)

func setupService(envPath, binPath, servicePath, veritasOpts string) error {
	opts := []byte(fmt.Sprintf("VERITAS_OPTS=%s", veritasOpts))

	if err := os.WriteFile(envPath, opts, 0o600); err != nil {
		log.Error("opts are not written successfully to env file")
		return err
	}

	serviceIni := fmt.Sprintf(
		`
		[Unit]
		Description=Veritas Container Registry

		[Service]
		EnvironmentFile=%s
		ExecStart=%s $VERITAS_OPTS
		Restart=on-failure

		[Install]
		WantedBy=default.target
		`,
		envPath,
		binPath,
	)

	err := os.WriteFile(servicePath, []byte(serviceIni), 0o600)
	if err != nil {
		log.Error("service file was not written successfully")
		return err
	}

	return nil
}
