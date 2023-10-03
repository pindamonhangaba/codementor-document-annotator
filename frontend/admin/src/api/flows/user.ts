import { api } from "../clients";
import { state, setUser } from "../../repo/session";
import { setUserProfile } from "../../repo/user";
import { UserMe, UserUpdateForm } from "../types";

export function getUserMe(): Promise<UserMe> {
  return new Promise<UserMe>((resolve, reject) => {
    const { jwt, user } = state();
    if (!jwt) {
      reject("empty authentication token");
      return;
    }
    if (!user?.userID) {
      reject("empty user id");
      return;
    }
    const Authorization = `Bearer ${jwt}`;

    api
      .usersGetProfileMe({
        params: { userID: user?.userID },
        headers: { Authorization },
      })
      .then((r) => {
        const u = r.data.item;
        setUserProfile(u);
        resolve(u);
      })
      .catch(reject);
  });
}

export type UpdatedUser = Awaited<
  ReturnType<typeof api.usersPatchUserByID>
>["data"]["item"];

function setUpdatedUser(user: UpdatedUser) {
  setUser({
    userID: user.userID,
    applID: user.applID,
    avatar: user?.avatar ?? "",
    name: user.name,
  }),
    setUserProfile((p) => ({ ...p, ...user } as UserMe));
}

export function updateUser(form: UserUpdateForm): Promise<UpdatedUser> {
  return new Promise<UpdatedUser>((resolve, reject) => {
    const { jwt, user } = state();
    if (!jwt) {
      reject("empty authentication token");
      return;
    }
    if (!user?.userID) {
      reject("empty user id");
      return;
    }
    const Authorization = `Bearer ${jwt}`;
    api
      .usersPatchUserByID(form, {
        params: { userID: user?.userID },
        headers: { Authorization },
      })
      .then((r) => {
        const u = r.data.item;
        setUpdatedUser(u);
        resolve(u);
      })
      .catch(reject);
  });
}
