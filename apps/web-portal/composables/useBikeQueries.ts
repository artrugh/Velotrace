import { useBikesApi } from "./useApi";

export const useBikeQueries = () => {
  const api = useBikesApi();

  const fetchMarketplace = () =>
    api.GET("/bikes", {}).then((res) => {
      if (res.error) throw res.error;
      return res.data?.bikes ?? [];
    });

  const fetchBikeById = (id: string) =>
    api.GET("/bikes/{id}", { params: { path: { id } } }).then((res) => {
      if (res.error) throw res.error;
      return res.data;
    });

  const fetchMyBikes = () =>
    api.GET("/my/bikes", {}).then((res) => {
      if (res.error) throw res.error;
      return res.data?.bikes ?? [];
    });

  return {
    fetchMarketplace,
    fetchBikeById,
    fetchMyBikes,
  };
};
