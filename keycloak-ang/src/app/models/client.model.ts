export interface Client {
  id?: string;
  clientId: string;
  name?: string;
  description?: string;
  enabled?: boolean;
  protocol?: string;
  redirectUris?: string[];
  webOrigins?: string[];
}
