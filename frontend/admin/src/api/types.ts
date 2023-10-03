import { api } from "./clients";

export type GeoLocation = Awaited<
  ReturnType<typeof api.geoLocationList>
>["data"]["items"][0];

export type Payment = Awaited<
  ReturnType<typeof api.marketplacePayment>
>["data"]["item"];

export type OnboardingForm = Parameters<typeof api.authCreateAccount>[0];

export type AuthResponse = Awaited<
  ReturnType<typeof api.authCreateAccount>
>["data"]["item"];

export type UserMe = Awaited<
  ReturnType<typeof api.usersGetProfileMe>
>["data"]["item"];

export type UserUpdateForm = Parameters<typeof api.usersPatchUserByID>[0];
