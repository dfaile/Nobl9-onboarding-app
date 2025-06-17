// src/utils/validation.ts
// Utility functions for input validation

/**
 * Validates the appID (project name).
 * @param appID - The project name to validate.
 * @returns {boolean} - true if valid, false otherwise.
 */
export function validateAppID(appID: string): boolean {
  // Example validation: appID must be non-empty and alphanumeric
  return appID.trim().length > 0 && /^[a-zA-Z0-9]+$/.test(appID);
}

/**
 * Validates a comma-separated list of user IDs.
 * @param userIds - Comma-separated user IDs to validate.
 * @returns {boolean} - true if valid, false otherwise.
 */
export function validateUserIDs(userIds: string): boolean {
  // Example validation: user IDs must be non-empty and alphanumeric
  const ids = userIds.split(',').map(id => id.trim());
  return ids.every(id => id.length > 0 && /^[a-zA-Z0-9]+$/.test(id));
} 