import type { InfinityQuery } from './query.types';
import type { DataSourceInstanceSettings, DataSourceJsonData } from '@grafana/data';

//#region Config
export interface GlobalInfinityQuery {
  name: string;
  id: string;
  query: InfinityQuery;
}
export type AuthType = 'none' | 'basicAuth' | 'apiKey' | 'bearerToken' | 'oauthPassThru' | 'digestAuth' | 'aws' | 'azureBlob' | 'oauth2' | 'zcap';
export type OAuth2Type = 'client_credentials' | 'jwt' | 'others';
export type APIKeyType = 'header' | 'query';
export type OAuth2Props = {
  oauth2_type?: OAuth2Type;
  client_id?: string;
  email?: string;
  private_key_id?: string;
  subject?: string;
  token_url?: string;
  scopes?: string[];
  authStyle?: number;
};
export type AWSAuthProps = {
  authType?: 'keys';
  region?: string;
  service?: string;
};
export type InfinityReferenceData = { name: string; data: string };
export type ProxyType = 'none' | 'env' | 'url';
export interface InfinityOptions extends DataSourceJsonData {
  auth_method?: AuthType;
  apiKeyKey?: string;
  apiKeyType?: APIKeyType;
  oauth2?: OAuth2Props;
  aws?: AWSAuthProps;
  zcapJsonPath?: string; 
  tlsSkipVerify?: boolean;
  tlsAuth?: boolean;
  serverName?: string;
  tlsAuthWithCACert?: boolean;
  global_queries?: GlobalInfinityQuery[];
  timeoutInSeconds?: number;
  proxy_type?: ProxyType;
  proxy_url?: string;
  oauthPassThru?: boolean;
  allowedHosts?: string[];
  refData?: InfinityReferenceData[];
  enableOpenApi?: boolean;
  openApiVersion?: 'open-api-2.0';
  openApiUrl?: string;
  openAPIBaseURL?: string;
  customHealthCheckEnabled?: boolean;
  customHealthCheckUrl?: string;
  azureBlobAccountUrl?: string;
  azureBlobAccountName?: string;
}

export interface InfinitySecureOptions {
  basicAuthPassword?: string;
  tlsCACert?: string;
  tlsClientCert?: string;
  tlsClientKey?: string;
  apiKeyValue?: string;
  bearerToken?: string;
  awsAccessKey?: string;
  awsSecretKey?: string;
  oauth2ClientSecret?: string;
  oauth2JWTPrivateKey?: string;
  azureBlobAccountKey?: string;
}
export interface SecureField {
  id: string;
  name: string;
  value: string;
  configured: boolean;
}
export type InfinityInstanceSettings = DataSourceInstanceSettings<InfinityOptions>;
//#endregion
