// auth.guard.ts
import { inject } from '@angular/core';
import { Router, type CanActivateFn } from '@angular/router';
import { KeycloakService } from 'keycloak-angular';

export const authGuard: CanActivateFn = async () => {
  const keycloak = inject(KeycloakService);
  const router = inject(Router);

  const isLoggedIn = await keycloak.isLoggedIn();
  
  if (isLoggedIn) {
    return true;
  } else {
    // Перенаправляем на страницу логина
    router.navigate(['/login']);
    return false;
  }
};