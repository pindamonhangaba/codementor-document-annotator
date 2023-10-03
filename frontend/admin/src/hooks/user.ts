import React from "react";
import { getUserMe } from "@/api/flows/user";
import { useUserStore } from "@/repo/user";
import type { UserMe } from "@/api/types";

export const useGetUserMe = () => {
  const { userProfile } = useUserStore();
  const [loading, setLoading] = React.useState(false);
  const [user, setUser] = React.useState<UserMe>();

  React.useEffect(() => {
    if (!userProfile) {
      setLoading(true);
      getUserMe()
        .catch(console.warn)
        .finally(() => setLoading(false));
    }
  }, [userProfile]);

  React.useEffect(() => {
    setUser(userProfile);
  }, [userProfile]);

  return [user, loading] as [typeof user, typeof loading];
};
