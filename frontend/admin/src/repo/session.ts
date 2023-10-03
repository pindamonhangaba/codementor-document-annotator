import { createStore } from 'zustand/vanilla';
import { useStore } from 'zustand';
import { getCookie } from '../cookie';
import jwt_decode from 'jwt-decode';

export const COOKIE_TOKEN_API = 'session_jwt';
export const COOKIE_EXPIRY = 5 * 60 * 1000;

export type SessionUser = {
  applID: string;
  avatar: string;
  name: string;
  userID: string;
};

type InitialState = {
  jwt: null | string;
  user?: SessionUser;
  payload: {
    exp: number;
    userID: string;
  } | null;
};

const store = createStore<InitialState>(() => ({
  jwt: getCookie(COOKIE_TOKEN_API) as null | string,
  user: undefined as SessionUser | undefined,
  payload: (getCookie(COOKIE_TOKEN_API)
    ? jwt_decode(getCookie(COOKIE_TOKEN_API)!)
    : null) as {
    exp: number;
    userID: string;
  } | null,
}));

const { setState, getState } = store;

export const setAuthSession = (jwt: string, user: SessionUser | undefined) => {
  const payload: any = jwt ? jwt_decode(jwt) : '';
  setState({ jwt, user, payload });
};

export const setUser = (user: SessionUser) => {
  setState({ user });
};

export const getToken = () => {
  const current = getState();
  return {
    token: current.jwt,
    tokenMedia: current.media,
  };
};

export function hydrate(props: Partial<InitialState>) {
  const current = getState();
  setState({ ...current, ...props });
}

export function state() {
  return getState();
}

export const useSessionStore = () => useStore(store, (s) => s);
