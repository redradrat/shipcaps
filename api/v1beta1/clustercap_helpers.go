package v1beta1

func (cap *ClusterCap) GetSpec() CapSpec {
	return cap.Spec
}
