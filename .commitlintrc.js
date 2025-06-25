module.exports = {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "type-enum": [
      2,
      "always",
      [
        "feat", // New feature
        "fix", // Bug fix
        "docs", // Documentation changes
        "style", // Code style changes
        "refactor", // Code refactoring
        "test", // Test-related changes
        "chore", // Maintenance tasks
        "ci", // CI/CD changes
        "perf", // Performance improvements
        "build", // Build system changes
        "revert", // Revert a commit
        "wip", // Work in progress
        "security", // Security-related changes
        "ux", // User experience changes
        "ui", // User interface changes
      ],
    ],
    "subject-case": [2, "always", "sentence-case"], // Enforce sentence case for the subject
  },
};
