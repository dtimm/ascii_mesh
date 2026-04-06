# ASCII Mesh tool

# Technical Practices
- Use test-driven development! Red, green, refactor.
- 100% unit test coverage for new code
- Never use regex to parse HTML
- The development environment should mirror production
- Address issues directly - don't work around them
- Review and refactor code regularly
- Master fundamentals before adding complexity
- Make one change at a time
- When there are multiple possible changes, ask which one should be made first

## Testing Principles

1. Test-First Development
   - Write one test at a time
   - Never proceed past failing tests
   - Start with the simplest solution that could work
   - Either write a test OR make a failing test pass - never both
   - A unit test should make just one assertion
   - Don't make any changes to the application code until we see the test fail
   - After a failed implementation attempt:
     1. Use `git restore` to revert all code changes
     2. Re-run the failing test to confirm we're back to the original failure
     3. Think through a simpler solution before making any new changes
     4. If no simpler solution is apparent, ask for help rather than trying increasingly complex fixes

2. Test Quality
   - No sleeping in tests - use polling/waiting mechanisms
   - Keep test data simple and shared across languages
   - Use identical test data when testing across multiple languages
   - Use the fixtures in setup_test_fixtures.ex to setup data for tests
   
3. Test Process
   - Make one test pass completely before moving on
   - Run full test suite before suggesting improvements
   - Maintain high test coverage as a defense against regression

4. Debugging Test Failures
   - Analyze all moving parts before making any changes
   - After 2 failed fix attempts, pause and explain system understanding
   - Map the data flow from database to UI before fixing UI-related failures
   - For each failure, explain what the test verifies and how the system achieves that goal

## Decision Making

1. Trust your analysis:
   - If your technical assessment suggests a test is needed, explain why
   - Don't change your position just because the user initially disagrees
   - Support your position with specific technical reasoning
2. When disagreeing with the user:
   - Acknowledge their perspective
   - Explain your reasoning clearly
   - Suggest running the test to verify
3. Remember: It's better to have a test that passes immediately than to miss a potential regression

## Shell Command Guidelines

- Verify if user is working on Linux, Windows or MacOS
- On windows: commands in PowerShell by default
- On Linux: always start with `set -euo pipefail`
- Must pass shellcheck standards
- Prefer pipe-based solutions over control flow
- If control flow (if/while) seems necessary, suggest using another language
- Pipeline-oriented solutions are preferred

# Core Development Philosophy
- Test-Driven Development is our foundation
- Move fast with technical excellence
- Focus obsessively on working software
- "Do Right Longer" - maintain discipline through the entire project
- Make one change at a time, always verify it works
