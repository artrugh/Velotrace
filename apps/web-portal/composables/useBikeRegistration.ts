import type { Bike } from "./useApi";
import type { components } from "@api-contract/.generated/bikes";

export type BikeRegistrationRequest =
  components["schemas"]["handler.RegisterBikeRequest"];

export const useBikeRegistration = () => {
  const api = useBikesApi();

  const isRegistering = ref(false);
  const registrationError = ref<string | null>(null);
  const registrationProgress = ref(0);

  const registerBike = async (
    form: BikeRegistrationRequest,
    images: File[],
  ): Promise<Bike> => {
    isRegistering.value = true;
    registrationError.value = null;
    registrationProgress.value = 0;

    try {
      const { data: bike, error: apiError } = await api.POST("/bikes", {
        body: form,
      });

      if (apiError || !bike) {
        throw new Error(
          (apiError as any)?.error || "Failed to create bike record.",
        );
      }

      const bikeId = bike.id!;
      registrationProgress.value = 10;

      if (images.length === 0) {
        registrationProgress.value = 100;
        return bike;
      }

      const stepWeight = 90 / (images.length * 3);

      const processImage = async (file: File) => {
        const { data: urlData, error: urlError } = await api.POST(
          "/bikes/{id}/upload-url",
          {
            params: { path: { id: bikeId } },
            body: { filename: file.name },
          },
        );

        if (urlError || !urlData) {
          throw new Error(
            (urlError as any)?.error ||
              `Could not get upload URL for ${file.name}`,
          );
        }
        registrationProgress.value += stepWeight;

        const { upload_url, object_key } = urlData;

        const uploadResponse = await fetch(upload_url!, {
          method: "PUT",
          body: file,
          headers: { "Content-Type": file.type },
        });

        if (!uploadResponse.ok) {
          throw new Error(
            `Failed to upload ${file.name}. Storage returned ${uploadResponse.status}.`,
          );
        }
        registrationProgress.value += stepWeight;

        const { error: confirmError } = await api.POST(
          "/bikes/{id}/images/confirm",
          {
            params: { path: { id: bikeId } },
            body: { object_key: object_key! },
          },
        );

        if (confirmError) {
          throw new Error(
            (confirmError as any)?.error || `Failed to finalize ${file.name}`,
          );
        }
        registrationProgress.value += stepWeight;

        return object_key;
      };

      await Promise.all(images.map(processImage));

      registrationProgress.value = 100;
      return bike;
    } catch (e: any) {
      registrationError.value = e.message || "An unexpected error occurred.";
      throw e;
    } finally {
      isRegistering.value = false;
    }
  };

  return {
    registerBike,
    isRegistering,
    registrationError,
    registrationProgress,
  };
};
