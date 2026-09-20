package testdata

func goodNestedBlock() error {
	{
		err := example()
		if err != nil {
			return err
		}
	}

	return nil
}
