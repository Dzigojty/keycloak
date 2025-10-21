// src/app/services/keycloak.service.ts
import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable, BehaviorSubject, from } from 'rxjs';
import { tap, catchError, switchMap } from 'rxjs/operators';

// Keycloak конфигурация - обновите под ваш контейнер
const keycloakConfig = {
  url: 'http://localhost:8080',
  realm: 'my-app',                    // Имя вашего realm
  clientId: 'angular-app',             // Client ID который вы создали
  username: 'admin',
  password: 'admin'
};

@Injectable({
  providedIn: 'root'
})
export class KeycloakAdminService {
  private isAuthenticatedSubject = new BehaviorSubject<boolean>(false);
  private userProfileSubject = new BehaviorSubject<any>(null);
  
  public isAuthenticated$ = this.isAuthenticatedSubject.asObservable();
  public userProfile$ = this.userProfileSubject.asObservable();

  constructor(private http: HttpClient) {
    this.checkAuthStatus();
  }

  /**
   * Стандартный вход через редирект на страницу Keycloak
   */
  login(): void {
    const keycloakLoginUrl = `${keycloakConfig.url}/realms/${keycloakConfig.realm}/protocol/openid-connect/auth` +
      `?client_id=${keycloakConfig.clientId}` +
      `&redirect_uri=${encodeURIComponent(window.location.origin)}` +
      `&response_type=code` +
      `&scope=openid`;
    
    console.log('Redirecting to Keycloak:', keycloakLoginUrl);
    window.location.href = keycloakLoginUrl;
  }

  /**
   * Прямой вход (Resource Owner Password Credentials Grant)
   * Должен быть включен в настройках клиента Keycloak
   */
  loginDirect(username: string, password: string): Observable<any> {
    const body = new URLSearchParams();
    body.set('client_id', keycloakConfig.clientId);
    body.set('username', username);
    body.set('password', password);
    body.set('grant_type', 'password');
    body.set('scope', 'openid');

    const headers = new HttpHeaders({
      'Content-Type': 'application/x-www-form-urlencoded'
    });

    return this.http.post<any>(
      `${keycloakConfig.url}/realms/${keycloakConfig.realm}/protocol/openid-connect/token`,
      body.toString(),
      { headers }
    ).pipe(
      tap(response => {
        console.log('Keycloak login successful:', response);
        this.handleAuthSuccess(response);
      }),
      catchError(error => {
        console.error('Keycloak login error:', error);
        throw error;
      })
    );
  }

  /**
   * Обработка успешной аутентификации
   */
  private handleAuthSuccess(response: any): void {
    localStorage.setItem('access_token', response.access_token);
    localStorage.setItem('refresh_token', response.refresh_token);
    localStorage.setItem('token_type', response.token_type);
    
    // Декодируем токен чтобы получить информацию о пользователе
    const tokenPayload = this.parseJwt(response.access_token);
    this.userProfileSubject.next(tokenPayload);
    this.isAuthenticatedSubject.next(true);
  }

  /**
   * Парсинг JWT токена
   */
  private parseJwt(token: string): any {
    try {
      const base64Url = token.split('.')[1];
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
      }).join(''));

      return JSON.parse(jsonPayload);
    } catch (e) {
      console.error('Error parsing JWT:', e);
      return null;
    }
  }

  /**
   * Выход
   */
  logout(): void {
    const token = this.getToken();
    localStorage.clear();
    this.isAuthenticatedSubject.next(false);
    this.userProfileSubject.next(null);
    
    const logoutUrl = `${keycloakConfig.url}/realms/${keycloakConfig.realm}/protocol/openid-connect/logout` +
      `?post_logout_redirect_uri=${encodeURIComponent(window.location.origin)}`;
    
    window.location.href = logoutUrl;
  }

  /**
   * Регистрация нового пользователя
   */
  register(): void {
    const registerUrl = `${keycloakConfig.url}/realms/${keycloakConfig.realm}/protocol/openid-connect/registrations` +
      `?client_id=${keycloakConfig.clientId}` +
      `&redirect_uri=${encodeURIComponent(window.location.origin)}` +
      `&response_type=code` +
      `&scope=openid`;
    
    window.location.href = registerUrl;
  }

  /**
   * Восстановление пароля
   */
  forgotPassword(): void {
    const resetUrl = `${keycloakConfig.url}/realms/${keycloakConfig.realm}/login-actions/reset-credentials` +
      `?client_id=${keycloakConfig.clientId}`;
    
    window.location.href = resetUrl;
  }

  /**
   * Проверка статуса аутентификации
   */
  isLoggedIn(): boolean {
    return this.isAuthenticatedSubject.value;
  }

  /**
   * Получение токена
   */
  getToken(): string | null {
    return localStorage.getItem('access_token');
  }

  /**
   * Получение информации о пользователе
   */
  getUser(): any {
    return this.userProfileSubject.value;
  }

  /**
   * Проверка роли пользователя
   */
  hasRole(role: string): boolean {
    const user = this.getUser();
    return user && user.realm_access && user.realm_access.roles && user.realm_access.roles.includes(role);
  }

  /**
   * Проверка статуса аутентификации при загрузке приложения
   */
  private checkAuthStatus(): void {
    // Проверяем URL на наличие authorization code (после редиректа от Keycloak)
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');
    
    if (code) {
      this.exchangeCodeForToken(code);
    } else {
      // Проверяем наличие токена в localStorage
      const token = this.getToken();
      if (token && !this.isTokenExpired(token)) {
        const tokenPayload = this.parseJwt(token);
        this.isAuthenticatedSubject.next(true);
        this.userProfileSubject.next(tokenPayload);
      }
    }
  }

  /**
   * Проверка истечения срока действия токена
   */
  private isTokenExpired(token: string): boolean {
    try {
      const payload = this.parseJwt(token);
      if (!payload || !payload.exp) return true;
      
      return Date.now() >= payload.exp * 1000;
    } catch (e) {
      return true;
    }
  }

  /**
   * Обмен authorization code на access token
   */
  private exchangeCodeForToken(code: string): void {
    const body = new URLSearchParams();
    body.set('client_id', keycloakConfig.clientId);
    body.set('code', code);
    body.set('grant_type', 'authorization_code');
    body.set('redirect_uri', window.location.origin);

    const headers = new HttpHeaders({
      'Content-Type': 'application/x-www-form-urlencoded'
    });

    this.http.post<any>(
      `${keycloakConfig.url}/realms/${keycloakConfig.realm}/protocol/openid-connect/token`,
      body.toString(),
      { headers }
    ).subscribe({
      next: (response) => {
        this.handleAuthSuccess(response);
        
        // Очищаем URL от параметров
        window.history.replaceState({}, document.title, window.location.pathname);
      },
      error: (error) => {
        console.error('Token exchange error:', error);
      }
    });
  }

  /**
   * Обновление токена
   */
  refreshToken(): Observable<any> {
    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken) {
      throw new Error('No refresh token available');
    }

    const body = new URLSearchParams();
    body.set('client_id', keycloakConfig.clientId);
    body.set('grant_type', 'refresh_token');
    body.set('refresh_token', refreshToken);

    const headers = new HttpHeaders({
      'Content-Type': 'application/x-www-form-urlencoded'
    });

    return this.http.post<any>(
      `${keycloakConfig.url}/realms/${keycloakConfig.realm}/protocol/openid-connect/token`,
      body.toString(),
      { headers }
    ).pipe(
      tap(response => {
        this.handleAuthSuccess(response);
      })
    );
  }

  // Получение токена администратора
  private async getAdminToken(): Promise<string> {
    const body = new URLSearchParams();
    body.set('client_id', 'admin-cli');
    body.set('username', keycloakConfig.username);
    body.set('password', keycloakConfig.password);
    body.set('grant_type', 'password');

    const response = await fetch(`${keycloakConfig.url}/realms/master/protocol/openid-connect/token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: body.toString()
    });
    
    const data = await response.json();
    return data.access_token;
  }

  // Создание пользователя в Keycloak
  async createUser(userData: any): Promise<any> {
    const token = await this.getAdminToken();
    
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });

    const userPayload = {
      username: userData.username,
      email: userData.email,
      firstName: userData.firstName,
      lastName: userData.lastName,
      enabled: true,
      credentials: [{
        type: "password",
        value: userData.password,
        temporary: false
      }]
    };

    return this.http.post(
      `${keycloakConfig.url}/admin/realms/${keycloakConfig.realm}/users`,
      userPayload,
      { headers }
    ).toPromise();
  }
}
