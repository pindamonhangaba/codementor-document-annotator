import { pluginToken } from '@zodios/plugins';
import { createApiClient } from './zodios/api';
import { eraseCookie, getCookie, setCookie } from '../cookie';
import {
  setAuthSession,
  COOKIE_TOKEN_API,
  COOKIE_EXPIRY,
} from '../repo/session';
export { schemas } from './zodios/api';

function newApiClient() {
  if (typeof window !== 'undefined') {
    return createApiClient('/');
  }
  let baseUrl = process.env.API_PATH?? 'http://localhost:8089';
  return createApiClient(baseUrl);
}


const api = newApiClient();



const after: ReturnType<typeof pluginToken> = {
  name: 'after',

  request: async (api, config) => {
    const conf = { ...config, withCredentials: true };
    return conf;
  },

  response: async (api, config, response) => {
    try {
      const {
        data: {
          item: { jwt, user },
        },
      } = response.data;
      setCookie(
        COOKIE_TOKEN_API,
        jwt,
        new Date(new Date().getTime() + COOKIE_EXPIRY)
      );
      setAuthSession(jwt, user);
    } catch (e) {
      console.warn(e);
    }
    return response;
  },
};

const afterLogout: ReturnType<typeof pluginToken> = {
  name: 'afterLogout',

  response: async (api, config, response) => {
    try {
      eraseCookie(COOKIE_TOKEN_API);
      setAuthSession('', undefined);
    } catch (e) {
      console.warn(e);
    }
    return response;
  },
};


if (typeof window !== 'undefined') {
  api.use('post', '/api/onboarding/create-account', after);
  api.use('post', '/api/auth/signin', after);
  api.use('get', '/api/auth/refresh', after);
  api.use('post', '/api/auth/signout', afterLogout);
  api.use(
    pluginToken({
      getToken: async () => getCookie(COOKIE_TOKEN_API) ?? undefined,
      renewToken: async () => {
        const {
          data: {
            item: { jwt },
          },
        } = await api.get('/api/auth/refresh');
        return jwt;
      },
    })
  );
}

function initSession() {
  api.authRefresh().catch(console.warn);
  setInterval(() => {
    api.authRefresh().catch(console.warn);
  }, COOKIE_EXPIRY - 5000);
}

initSession();

export { api };
