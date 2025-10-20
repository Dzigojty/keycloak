export interface User {
  id?: string;
  username: string;
  email: string;
  firstName?: string;
  lastName?: string;
  enabled?: boolean;
  credentials?: Credential[];
  attributes?: { [key: string]: any };
}

export interface Credential {
  type: string;
  value: string;
  temporary?: boolean;
}

// Создадим тип для создания пользователя с обязательными полями
export interface CreateUserRequest {
  username: string;
  password: string;
  position: string
}
