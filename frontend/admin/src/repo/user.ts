import { createStore } from "zustand/vanilla";
import { useStore } from "zustand";
import { UserMe } from "@/api/types";

type InitialState = {
  userProfile?: UserMe;
};

const store = createStore<InitialState>(() => ({
  userProfile: undefined,
}));

const { setState, getState } = store;

export const setUserProfile = (u: UserMe | ((u?: UserMe) => UserMe)) => {
  if (typeof u === "function") {
    const { userProfile } = getState();
    setState({ userProfile: u(userProfile) });
    return;
  }

  setState({ userProfile: u });
};

export function hydrate(props: Partial<InitialState>) {
  const current = getState();
  setState({ ...current, ...props });
}

export function state() {
  return getState();
}

export const useUserStore = () => useStore(store, (s) => s);
