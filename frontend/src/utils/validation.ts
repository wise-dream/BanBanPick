/**
 * Централизованная валидация форм
 */

export interface ValidationRule {
  validate: (value: any) => boolean;
  message: string;
}

export interface ValidationResult {
  isValid: boolean;
  errors: Record<string, string>;
}

/**
 * Валидация email
 */
export function validateEmail(email: string): boolean {
  if (!email || !email.trim()) {
    return false;
  }
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email.trim());
}

/**
 * Валидация username
 */
export function validateUsername(username: string): { isValid: boolean; error?: string } {
  if (!username || !username.trim()) {
    return { isValid: false, error: 'Имя пользователя обязательно' };
  }

  const trimmed = username.trim();

  if (trimmed.length < 3) {
    return { isValid: false, error: 'Имя пользователя должно быть не менее 3 символов' };
  }

  if (trimmed.length > 50) {
    return { isValid: false, error: 'Имя пользователя должно быть не более 50 символов' };
  }

  // Разрешаем только буквы, цифры, подчеркивания и дефисы
  const usernameRegex = /^[a-zA-Z0-9_-]+$/;
  if (!usernameRegex.test(trimmed)) {
    return { isValid: false, error: 'Имя пользователя может содержать только буквы, цифры, подчеркивания и дефисы' };
  }

  return { isValid: true };
}

/**
 * Валидация password
 */
export function validatePassword(password: string): { isValid: boolean; error?: string } {
  if (!password) {
    return { isValid: false, error: 'Пароль обязателен' };
  }

  if (password.length < 6) {
    return { isValid: false, error: 'Пароль должен быть не менее 6 символов' };
  }

  if (password.length > 128) {
    return { isValid: false, error: 'Пароль должен быть не более 128 символов' };
  }

  return { isValid: true };
}

/**
 * Валидация кода комнаты
 */
export function validateRoomCode(code: string): { isValid: boolean; error?: string } {
  if (!code || !code.trim()) {
    return { isValid: false, error: 'Код комнаты обязателен' };
  }

  const trimmed = code.trim().toUpperCase();

  // Код должен быть 6-8 символов, буквы и цифры
  if (trimmed.length < 6 || trimmed.length > 8) {
    return { isValid: false, error: 'Код комнаты должен быть от 6 до 8 символов' };
  }

  const codeRegex = /^[A-Z0-9]+$/;
  if (!codeRegex.test(trimmed)) {
    return { isValid: false, error: 'Код комнаты может содержать только буквы и цифры' };
  }

  return { isValid: true };
}

/**
 * Валидация названия комнаты
 */
export function validateRoomName(name: string): { isValid: boolean; error?: string } {
  if (!name || !name.trim()) {
    return { isValid: false, error: 'Название комнаты обязательно' };
  }

  const trimmed = name.trim();

  if (trimmed.length < 3) {
    return { isValid: false, error: 'Название комнаты должно быть не менее 3 символов' };
  }

  if (trimmed.length > 50) {
    return { isValid: false, error: 'Название комнаты должно быть не более 50 символов' };
  }

  return { isValid: true };
}

/**
 * Валидация пароля комнаты (опциональный)
 */
export function validateRoomPassword(password: string | null | undefined): { isValid: boolean; error?: string } {
  // Пароль опционален
  if (!password || !password.trim()) {
    return { isValid: true };
  }

  if (password.length < 4) {
    return { isValid: false, error: 'Пароль комнаты должен быть не менее 4 символов' };
  }

  if (password.length > 50) {
    return { isValid: false, error: 'Пароль комнаты должен быть не более 50 символов' };
  }

  return { isValid: true };
}

/**
 * Валидация формы регистрации
 */
export function validateRegisterForm(data: {
  email: string;
  username: string;
  password: string;
  confirmPassword: string;
}): ValidationResult {
  const errors: Record<string, string> = {};

  // Email
  if (!data.email || !data.email.trim()) {
    errors.email = 'Email обязателен';
  } else if (!validateEmail(data.email)) {
    errors.email = 'Некорректный формат email';
  }

  // Username
  const usernameValidation = validateUsername(data.username);
  if (!usernameValidation.isValid) {
    errors.username = usernameValidation.error || 'Некорректное имя пользователя';
  }

  // Password
  const passwordValidation = validatePassword(data.password);
  if (!passwordValidation.isValid) {
    errors.password = passwordValidation.error || 'Некорректный пароль';
  }

  // Confirm Password
  if (data.password !== data.confirmPassword) {
    errors.confirmPassword = 'Пароли не совпадают';
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

/**
 * Валидация формы входа
 */
export function validateLoginForm(data: {
  email: string;
  password: string;
}): ValidationResult {
  const errors: Record<string, string> = {};

  // Email
  if (!data.email || !data.email.trim()) {
    errors.email = 'Email обязателен';
  } else if (!validateEmail(data.email)) {
    errors.email = 'Некорректный формат email';
  }

  // Password
  if (!data.password) {
    errors.password = 'Пароль обязателен';
  } else if (data.password.length < 6) {
    errors.password = 'Пароль должен быть не менее 6 символов';
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}

/**
 * Валидация формы создания комнаты
 */
export function validateCreateRoomForm(data: {
  name: string;
  password?: string | null;
}): ValidationResult {
  const errors: Record<string, string> = {};

  // Name
  const nameValidation = validateRoomName(data.name);
  if (!nameValidation.isValid) {
    errors.name = nameValidation.error || 'Некорректное название комнаты';
  }

  // Password (опциональный)
  const passwordValidation = validateRoomPassword(data.password);
  if (!passwordValidation.isValid) {
    errors.password = passwordValidation.error || 'Некорректный пароль';
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
}
