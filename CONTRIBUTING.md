# Contributing to Hera

Contributions are very much welcomed and appreciated. We want to make contributing to this project as easy and transparent as possible, so please read through and become familiar with these guidelines.

## Pull Requests

### We use [Github Flow](https://guides.github.com/introduction/flow/index.html) for pull requests
When submitting a pull request:

1. Fork the repo and create your branch from `master`.
2. Add tests to your changes whenever possible.
3. Update the documentation when changing any user-facing behavior.
4. Submit your PR

#### Continuous Integration

Hera uses [Semaphore](https://semaphoreci.com/aschaper/hera) to build pull requests. Ensure the build for your PR succeeds and that the tests pass.

## Reporting Bugs

Report your bug by creating a [new issue](https://github.com/aschaper/hera/issues) in this repository.

### Be Descriptive

**Great bug reports** tend to have:

- A quick summary and/or background
- Steps to reproduce
- Expected behavior
- Actual behavior
  * Include a sample output of your logs if any
- Notes (possibly including why you think this might be happening, or things you've tried that didn't work)

## Use a Consistent Coding Style

Run [golint](https://github.com/golang/lint) against your code until it reports no issues.

## License

By contributing, you agree that your contributions will be licensed under its [MIT License](https://github.com/aschaper/hera/blob/master/LICENSE).
