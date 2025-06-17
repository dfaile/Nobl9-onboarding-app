import { validateAppID, validateUserIDs } from '../../utils/validation';

describe('Validation Utilities', () => {
  test('validateAppID returns true for valid appID', () => {
    expect(validateAppID('MyApp123')).toBe(true);
  });

  test('validateAppID returns false for empty or invalid appID', () => {
    expect(validateAppID('')).toBe(false);
    expect(validateAppID(' ')).toBe(false);
    expect(validateAppID('My App!')).toBe(false);
  });

  test('validateUserIDs returns true for valid comma-separated user IDs', () => {
    expect(validateUserIDs('user1,user2,user3')).toBe(true);
    expect(validateUserIDs('user1, user2, user3')).toBe(true);
  });

  test('validateUserIDs returns false for empty or invalid user IDs', () => {
    expect(validateUserIDs('')).toBe(false);
    expect(validateUserIDs('user1, ,user3')).toBe(false);
    expect(validateUserIDs('user1, user!')).toBe(false);
  });
}); 