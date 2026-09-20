package testdata

func goodShadowedError() (err error) {
	{
		err := example()
		if err != nil {
			return err
		}
	}

	return nil
}