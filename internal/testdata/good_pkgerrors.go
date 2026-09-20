package testdata

import (
	pkgerrors "github.com/pkg/errors"
)

func goodPkgErrorsErrorf() error {
	err := pkgerrors.Errorf("something bad happened")
	sink(err)
	return err
}

func goodPkgErrorsNew() error {
	err := pkgerrors.New("something bad happened")
	sink(err)
	return err
}

func goodPkgErrorsWithMessage() error {
	if err := example(); err != nil {
		mErr := pkgerrors.WithMessage(err, "something bad happened")
		sink(err)
		return mErr
	}
	return nil
}

func goodPkgErrorsWithMessagef() error {
	if err := example(); err != nil {
		mErr := pkgerrors.WithMessagef(err, "something bad happened, %d", 1)
		sink(mErr)
		return mErr
	}
	return nil
}

func goodPkgErrorsWithStack() error {
	if err := example(); err != nil {
		mErr := pkgerrors.WithStack(err)
		sink(mErr)
		return mErr
	}
	return nil
}

func goodPkgErrorsWrap() error {
	if err := example(); err != nil {
		mErr := pkgerrors.Wrap(err, "wrapped")
		sink(err)
		return mErr
	}
	return nil
}

func goodPkgErrorsWrapf() error {
	if err := example(); err != nil {
		mErr := pkgerrors.Wrapf(err, "wrapped, %d", 1)
		sink(err)
		return mErr
	}
	return nil
}
